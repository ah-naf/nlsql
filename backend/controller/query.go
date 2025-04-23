package controller

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var destructiveRE = regexp.MustCompile(`(?i)^\s*(INSERT|UPDATE|DELETE|DROP|ALTER|CREATE)`)

const deepseekURL = "https://api.together.xyz/v1/chat/completions"
const MODEL_NAME = "meta-llama/Llama-4-Maverick-17B-128E-Instruct-FP8"

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
}

func HandleNLQuery(c *gin.Context) {
	var req struct {
		Config DBRequest `json:"config"`
		Prompt string    `json:"prompt"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
		return
	}

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		req.Config.Host, req.Config.Port, req.Config.User, req.Config.Pass, req.Config.DBName,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Connection error: " + err.Error()})
		return
	}
	defer db.Close()

	// STEP 1: Get all table names
	tableNames, err := getTables(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Table name error: " + err.Error()})
		return
	}

	// STEP 2: Ask LLM to detect relevant tables
	prompt := buildTableDetectionPrompt(tableNames, req.Prompt)
	llmMessages := []Message{
		{Role: "system", Content: "You are a helpful assistant that selects only relevant table names from a schema list."},
		{Role: "user", Content: prompt},
	}

	response, err := connectLLM(llmMessages)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "LLM error: " + err.Error()})
		return
	}

	// STEP 3: Parse LLM output into []string
	detectedTables := parseCSV(response) // e.g. "users, courses" → ["users", "courses"]

	// STEP 4: Load full schema
	fullSchema, err := loadFullSchema(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Schema load error: " + err.Error()})
		return
	}

	// STEP 5: Filter schema to only relevant tables
	relevantSchema := map[string]TableInfo{}
	for _, tbl := range detectedTables {
		if info, ok := fullSchema[strings.TrimSpace(tbl)]; ok {
			relevantSchema[tbl] = info
		}
	}

	// STEP 6: Generate prompt to ask LLM for SQL
	sqlPrompt := buildSQLPrompt(relevantSchema, req.Prompt)
	sqlMessages := []Message{
		{Role: "system", Content: "You are an expert SQL assistant. Convert natural language into SQL using the schema below."},
		{Role: "user", Content: sqlPrompt},
	}

	sqlResult, err := connectLLM(sqlMessages)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "LLM error during SQL generation: " + err.Error()})
		return
	}

	sqlQuery := strings.TrimSpace(sqlResult)

	// Step 5: Execute the SQL query
	rows, err := db.Query(sqlQuery)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "SQL execution failed",
			"sql":        sqlQuery,
			"raw_output": sqlResult,
			"llm_tables": detectedTables,
		})
		return
	}
	defer rows.Close()

	// Step 6: Convert query result into []map[string]interface{}
	columns, _ := rows.Columns()
	result := []map[string]interface{}{}

	for rows.Next() {
		columnPointers := make([]interface{}, len(columns))
		columnValues := make([]interface{}, len(columns))

		for i := range columnPointers {
			columnPointers[i] = &columnValues[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			continue
		}

		rowMap := make(map[string]interface{})
		for i, colName := range columns {
			val := columnValues[i]
			rowMap[colName] = val
		}

		result = append(result, rowMap)
	}

	c.JSON(http.StatusOK, gin.H{
		"sql":          sqlResult,
		"result_table": result,
	})
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
		Model:    MODEL_NAME,
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

func getTables(db *sql.DB) ([]string, error) {
	rows, err := db.Query(`
			SELECT table_name
			FROM information_schema.tables
			WHERE table_schema='public'
			ORDER BY table_name
		`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tables := []string{}
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}

	return tables, nil
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

func parseCSV(s string) []string {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, ".\n") // remove trailing dot or newline
	parts := strings.Split(s, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func buildSQLPrompt(schema map[string]TableInfo, userQuery string) string {
	var sb strings.Builder

	for table, info := range schema {
		sb.WriteString(fmt.Sprintf("Table: %s\n", table))
		for _, col := range info.Columns {
			sb.WriteString(fmt.Sprintf("- %s: %s\n", col.Name, col.DataType))
		}
		sb.WriteString("\n")
	}

	return fmt.Sprintf(`
You are a SQL expert. Given the schema and user request below, generate a valid SQL query.
Only return the SQL. Do not explain anything. Do not format the sql. Give it in raw text.

### Schema
%s

### Request
%s

### SQL
`, sb.String(), userQuery)
}

func buildTableDetectionPrompt(tableNames []string, query string) string {
	return fmt.Sprintf(`
You are a database schema assistant.

You are given a list of table names:

%s

Based on the user's request, return only the **relevant table names** from the list above. 
Do not include any descriptions or explanations. Only output the table names as a comma-separated list.

### Request
%s

### Output (comma-separated table names only)
`, strings.Join(tableNames, ", "), query)
}
