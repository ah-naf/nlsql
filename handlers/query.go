package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"nlsql/models"
	"os"
	"regexp"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var destructiveRE = regexp.MustCompile(`(?i)^\s*(INSERT|UPDATE|DELETE|DROP|ALTER|CREATE)`)

const deepseekURL = "https://api.together.xyz/v1/chat/completions"

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type RequestBody struct {
	NLQuery   string    `json:"nl_query"`
	Confirmed bool      `json:"confirmed"`
	History   []Message `json:"history"`
}

type ResponseBody struct {
	Status            string                   `json:"status,omitempty"`
	Error             string                   `json:"error,omitempty"`
	NeedsConfirmation bool                     `json:"needs_confirmation,omitempty"`
	SQLPreview        string                   `json:"sql_preview,omitempty"`
	Table             []map[string]interface{} `json:"table,omitempty"`
	Message           string                   `json:"message,omitempty"`
	History           []Message                `json:"history,omitempty"`
	Schema            map[string][]string      `json:"schema,omitempty"`
}

func needsConfirmation(sql string) bool {
	return destructiveRE.MatchString(strings.TrimSpace(sql))
}

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}
}

func ShowQueryPage(c *gin.Context) {
	sess := sessions.Default(c)

	connStr := sess.Get("connection_string").(string)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	schema, err := models.GetSchema(db)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.HTML(http.StatusOK, "query.html", gin.H{
		"Schema": schema,
	})
}

func buildPrompt(schema map[string][]string, userText string) string {
	var parts []string
	for table, cols := range schema {
		parts = append(parts, fmt.Sprintf("%s(%s)", table, strings.Join(cols, ", ")))
	}
	schemaDefs := strings.Join(parts, ", ")
	return fmt.Sprintf(
		"Here are the table schemas: %s.\n"+
			"Generate an SQL query for the following request:\n"+
			"%s\n\n"+
			"***Only output the SQL query, with no explanation or markdown formatting.***",
		schemaDefs, userText,
	)
}

func connectLLM(messages []Message) (string, error) {
	loadEnv()
	token := os.Getenv("LLM_API_KEY")
	if token == "" {
		return "", fmt.Errorf("LLM API key not found in environment")
	}

	reqBody := struct {
		Model    string    `json:"model"`
		Messages []Message `json:"messages"`
	}{
		Model:    "meta-llama/Llama-3.3-70B-Instruct-Turbo-Free",
		Messages: messages,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}
	req, _ := http.NewRequest("POST", deepseekURL, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error: %s", body)
	}

	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("no response from LLM")
	}
	return out.Choices[0].Message.Content, nil
}

// HandleNLQuery handles incoming NL queries, uses client-side history management,
// builds the LLM prompt (including schema), and executes or previews SQL.
func HandleNLQuery(c *gin.Context) {
	// 1) Parse request with history
	var req RequestBody
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ResponseBody{
			Error: "invalid JSON: " + err.Error(),
		})
		return
	}

	// 2) Load schema from session (still keep schema in session as it's not changing often)
	sess := sessions.Default(c)
	rawSchema, ok := sess.Get("schema").(string)
	if !ok || rawSchema == "" {
		c.JSON(http.StatusBadRequest, ResponseBody{
			Error: "no schema in session; please re-select your database",
		})
		return
	}

	// 3) Initialize history if empty
	history := req.History
	if len(history) == 0 {
		history = []Message{
			{Role: "system", Content: "You are a helpful assistant. Only output SQL."},
		}
	}

	connStr, ok := sess.Get("connection_string").(string)
	if !ok || connStr == "" {
		c.JSON(http.StatusBadRequest, ResponseBody{
			Error: "no database connection in session",
		})
		return
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ResponseBody{
			Error: fmt.Sprintf("DB connect error: %v", err),
		})
		return
	}
	defer db.Close()

	schema, err := models.GetSchema(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ResponseBody{
			Error: fmt.Sprintf("Schema fetch error: %v", err),
		})
		return
	}

	// 4) Build the full prompt (includes schema + user text)
	userPrompt := buildPrompt(schema, req.NLQuery)

	// Create a copy of history to work with
	updatedHistory := append([]Message{}, history...)
	updatedHistory = append(updatedHistory, Message{Role: "user", Content: userPrompt})

	// 5) Call the LLM with the accumulated history
	sqlCommand, err := connectLLM(updatedHistory)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ResponseBody{
			Error:   err.Error(),
			History: history, // Return original history on error
		})
		return
	}

	// 6) If destructive SQL and not yet confirmed, ask for confirmation
	if needsConfirmation(sqlCommand) && !req.Confirmed {
		c.JSON(http.StatusOK, ResponseBody{
			NeedsConfirmation: true,
			SQLPreview:        sqlCommand,
			History:           updatedHistory, // Return updated history with user's message
		})
		return
	}

	// 7) Append the assistant's SQL to history
	updatedHistory = append(updatedHistory, Message{Role: "assistant", Content: sqlCommand})

	// 9) Execute or query depending on SQL verb
	upper := strings.ToUpper(strings.TrimSpace(sqlCommand))
	if strings.HasPrefix(upper, "SELECT") {
		rows, err := db.Query(sqlCommand)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ResponseBody{
				Error:   fmt.Sprintf("query error: %v", err),
				History: updatedHistory,
			})
			return
		}
		defer rows.Close()

		cols, _ := rows.Columns()
		result := []map[string]interface{}{}

		for rows.Next() {
			values := make([]interface{}, len(cols))
			pointers := make([]interface{}, len(cols))
			for i := range values {
				pointers[i] = &values[i]
			}
			if err := rows.Scan(pointers...); err != nil {
				c.JSON(http.StatusInternalServerError, ResponseBody{
					Error:   fmt.Sprintf("scan error: %v", err),
					History: updatedHistory,
				})
				return
			}
			rowMap := make(map[string]interface{}, len(cols))
			for i, col := range cols {
				rowMap[col] = values[i]
			}
			result = append(result, rowMap)
		}

		c.JSON(http.StatusOK, ResponseBody{
			Status:     "ok",
			SQLPreview: sqlCommand,
			Table:      result,
			History:    updatedHistory,
		})
		return
	}

	// 10) Non‑SELECT statements: Exec and report affected rows
	res, err := db.Exec(sqlCommand)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ResponseBody{
			Error:   fmt.Sprintf("exec error: %v", err),
			History: updatedHistory,
		})
		return
	}
	affected, _ := res.RowsAffected()

	schemaTemp, err := models.GetSchema(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ResponseBody{
			Error:   fmt.Sprintf("schema refresh error: %v", err),
			History: updatedHistory,
		})
		return
	}
	fmt.Println("schema temp", schemaTemp)

	c.JSON(http.StatusOK, ResponseBody{
		Status:     "ok",
		SQLPreview: sqlCommand,
		Message:    fmt.Sprintf("Query OK, %d rows affected", affected),
		History:    updatedHistory,
		Schema:     schemaTemp,
	})
}
