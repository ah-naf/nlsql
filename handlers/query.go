package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
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
	Status            string                         `json:"status,omitempty"`
	Error             string                         `json:"error,omitempty"`
	NeedsConfirmation bool                           `json:"needs_confirmation,omitempty"`
	SQLPreview        string                         `json:"sql_preview,omitempty"`
	Table             []map[string]interface{}       `json:"table,omitempty"`
	Message           string                         `json:"message,omitempty"`
	History           []Message                      `json:"history,omitempty"`
	Schema            map[string][]models.ColumnInfo `json:"schema,omitempty"`
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

	schemaBytes, _ := json.Marshal(schema)
	schemaJS := template.JS(schemaBytes)
	dbName := sess.Get("dbname").(string)

	c.HTML(http.StatusOK, "query.html", gin.H{
		"Schema":     schema,
		"SchemaJSON": schemaJS,
		"DBName":     dbName,
	})
}

func buildPrompt(schema map[string][]models.ColumnInfo, userText string) string {
	var parts []string

	for table, cols := range schema {
		// build column definitions with type (+ FK if any)
		var colDefs []string
		for _, c := range cols {
			def := fmt.Sprintf("%s %s", c.Name, c.DataType)
			if c.ForeignTable.Valid && c.ForeignColumn.Valid {
				def = fmt.Sprintf(
					"%s %s REFERENCES %s(%s)",
					c.Name,
					c.DataType,
					c.ForeignTable.String,
					c.ForeignColumn.String,
				)
			}
			colDefs = append(colDefs, def)
		}

		// join columns and wrap in TableName(...)
		parts = append(parts,
			fmt.Sprintf("%s(%s)", table, strings.Join(colDefs, ", ")),
		)
	}

	schemaDefs := strings.Join(parts, "; ")

	return fmt.Sprintf(
		"Here are the table schemas, including data types and foreign keys: %s.\n"+
			"Generate an SQL query for the following request:\n"+
			"%s\n\n"+
			"***Only output the SQL query, with no explanation or markdown formatting.***",
		schemaDefs,
		userText,
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

func HandleNLQuery(c *gin.Context) {
	var req RequestBody
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ResponseBody{
			Error: "invalid JSON: " + err.Error(),
		})
		return
	}

	sess := sessions.Default(c)

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

	userPrompt := buildPrompt(schema, req.NLQuery)

	updatedHistory := append([]Message{}, history...)
	updatedHistory = append(updatedHistory, Message{Role: "user", Content: userPrompt})

	sqlCommand, err := connectLLM(updatedHistory)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ResponseBody{
			Error:   err.Error(),
			History: history, // Return original history on error
		})
		return
	}

	if needsConfirmation(sqlCommand) && !req.Confirmed {
		updatedHistory = updatedHistory[:len(updatedHistory)-1]
		c.JSON(http.StatusOK, ResponseBody{
			NeedsConfirmation: true,
			SQLPreview:        sqlCommand,
			History:           updatedHistory, // Return updated history with user's message
		})
		return
	}

	updatedHistory = updatedHistory[:len(updatedHistory)-1]
	updatedHistory = append(updatedHistory, Message{Role: "assistant", Content: sqlCommand})

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
