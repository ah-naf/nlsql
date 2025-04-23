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
	Config       DBRequest `json:"config"`
	Prompt       string    `json:"prompt"`
	Confirmed    bool      `json:"confirmed"`
	SQLToConfirm string    `json:"sqlToConfirm"`
}

func HandleNLQuery(c *gin.Context) {
	var req RequestBody
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

	// 1) Fetch table names and detect relevant ones
	tableNames, err := getTables(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Table name error: " + err.Error()})
		return
	}
	detectorPrompt := buildTableDetectionPrompt(tableNames, req.Prompt)
	detectorMsgs := []Message{
		{Role: "system", Content: "You are a helpful assistant that selects only relevant table names from a schema list. If no relevant schema is found send '!!' as output."},
		{Role: "user", Content: detectorPrompt},
	}

	detectorResp, err := connectLLM(detectorMsgs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "LLM error: " + err.Error()})
		return
	}

	// prepare a variable to hold the final SQL we will execute
	var sqlQuery string

	// 2) Fallback QA‐wrapped path
	if strings.TrimSpace(detectorResp) == "!!" {
		qaMsgs := []Message{
			{
				Role: "system",
				Content: `You are a friendly assistant.  
When answering, return your response *only* as a single SQL SELECT statement that returns your text as a column named "output".  
For example:  
  SELECT 'Hi, I am doing fine' AS output;`,
			},
			{Role: "user", Content: req.Prompt},
		}
		ansSQL, err := connectLLM(qaMsgs)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "LLM QA error: " + err.Error()})
			return
		}
		sqlQuery = strings.TrimSpace(ansSQL)

	} else {
		// 3) Normal NL→SQL pipeline

		// parse the comma‐list into []string
		detectedTables := parseCSV(detectorResp)

		// load the full schema
		fullSchema, err := loadFullSchema(db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Schema load error: " + err.Error()})
			return
		}

		// filter to only the detected tables
		relevantSchema := map[string]TableInfo{}
		for _, tbl := range detectedTables {
			if info, ok := fullSchema[strings.TrimSpace(tbl)]; ok {
				relevantSchema[tbl] = info
			}
		}

		// build and call the SQL‐generation LLM
		sqlPrompt := buildSQLPrompt(relevantSchema, req.Prompt)
		sqlMsgs := []Message{
			{Role: "system", Content: "You are an expert SQL assistant. Convert natural language into SQL using the schema below."},
			{Role: "user", Content: sqlPrompt},
		}
		llmSQL, err := connectLLM(sqlMsgs)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "LLM error during SQL generation: " + err.Error()})
			return
		}
		sqlQuery = strings.TrimSpace(llmSQL)
	}

	// 4) Destructive check
	if needsConfirmation(sqlQuery) && !req.Confirmed {
		c.JSON(http.StatusOK, gin.H{
			"needs_confirmation": true,
			"sql_preview":        sqlQuery,
			"message":            "This query may modify your database. Please confirm before execution.",
		})
		return
	}

	// 5) Execute and return result_table
	rows, err := db.Query(sqlQuery)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "SQL execution failed",
			"sql":   sqlQuery,
		})
		return
	}
	defer rows.Close()

	cols, _ := rows.Columns()
	result := []map[string]interface{}{}
	for rows.Next() {
		ptrs := make([]interface{}, len(cols))
		vals := make([]interface{}, len(cols))
		for i := range ptrs {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			continue
		}
		row := map[string]interface{}{}
		for i, name := range cols {
			row[name] = vals[i]
		}
		result = append(result, row)
	}

	c.JSON(http.StatusOK, gin.H{
		"sql":          sqlQuery,
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

func getOperationType(sql string) string {
	sql = strings.TrimSpace(strings.ToUpper(sql))

	if strings.HasPrefix(sql, "INSERT") {
		return "INSERT operation"
	} else if strings.HasPrefix(sql, "UPDATE") {
		return "UPDATE operation"
	} else if strings.HasPrefix(sql, "DELETE") {
		return "DELETE operation"
	} else if strings.HasPrefix(sql, "DROP") {
		return "DROP operation"
	} else if strings.HasPrefix(sql, "ALTER") {
		return "ALTER operation"
	} else if strings.HasPrefix(sql, "CREATE") {
		return "CREATE operation"
	}

	return "SQL operation"
}
