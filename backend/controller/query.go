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
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var destructiveRE = regexp.MustCompile(`(?i)^\s*(INSERT|UPDATE|DELETE|DROP|ALTER|CREATE)`)
var dbOperationRE = regexp.MustCompile(`(?i)\b(table|database|column|row|insert|update|delete|drop|alter|create|select|from|where|join|schema)\b`)

const deepseekURL = "https://api.together.xyz/v1/chat/completions"
const MODEL_NAME = "meta-llama/Llama-4-Maverick-17B-128E-Instruct-FP8"
const MAX_HISTORY_ITEMS = 10            // Keep track of last 10 interactions
const HISTORY_EXPIRY = 30 * time.Minute // Session expires after 30 minutes of inactivity

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type HistoryItem struct {
	Prompt   string    `json:"prompt"`
	SQL      string    `json:"sql"`
	Response string    `json:"response"`
	Time     time.Time `json:"time"`
}

type ConversationHistory struct {
	Items    []HistoryItem `json:"items"`
	ClientIP string        `json:"client_ip"`
	LastUsed time.Time     `json:"last_used"`
}

type RequestBody struct {
	Config       DBRequest `json:"config"`
	Prompt       string    `json:"prompt"`
	Confirmed    bool      `json:"confirmed"`
	SQLToConfirm string    `json:"sqlToConfirm"`
	SessionID    string    `json:"sessionId"` // Optional session ID for tracking conversations
}

var (
	conversations     = make(map[string]*ConversationHistory)
	conversationMutex sync.Mutex
)

// Cleanup expired conversations periodically
func init() {
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			cleanupExpiredConversations()
		}
	}()
}

func cleanupExpiredConversations() {
	conversationMutex.Lock()
	defer conversationMutex.Unlock()

	now := time.Now()
	for id, history := range conversations {
		if now.Sub(history.LastUsed) > HISTORY_EXPIRY {
			delete(conversations, id)
		}
	}
}

func HandleNLQuery(c *gin.Context) {
	var req RequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
		return
	}

	// Generate session ID if not provided
	sessionID := req.SessionID
	if sessionID == "" {
		sessionID = fmt.Sprintf("%s-%s", c.ClientIP(), req.Config.DBName)
	}

	// Load or create conversation history
	history := getConversationHistory(sessionID, c.ClientIP())

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

	// Pre-check: Is this likely a DB operation request?
	isLikelyDBOperation := dbOperationRE.MatchString(req.Prompt)

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

	// 2) Check if this is a table detection miss
	// Return error if table detector returned "!!" and it's not likely a DB operation
	if strings.TrimSpace(detectorResp) == "!!" && !isLikelyDBOperation {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Could not find relevant tables for your query. Please rephrase with specific table references.",
			"session_id": sessionID,
		})
		return
	}

	// 3) Normal NL→SQL pipeline with conversation history

	// Build history context for the LLM
	var historyContext string
	if len(history.Items) > 0 {
		historyContext = "Previous related operations:\n\n"
		for i, item := range history.Items {
			historyContext += fmt.Sprintf("User: %s\nExecuted SQL: %s\n\n", item.Prompt, item.SQL)
			if i >= 4 { // Only include last 5 interactions for context
				break
			}
		}
	}

	// For CREATE/DROP/ALTER operations with no specific tables mentioned
	if isLikelyDBOperation && (strings.TrimSpace(detectorResp) == "!!" || strings.Contains(strings.ToLower(req.Prompt), "table")) {
		// Load the full schema
		fullSchema, err := loadFullSchema(db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Schema load error: " + err.Error()})
			return
		}

		sqlPrompt := buildSQLPromptWithHistory(fullSchema, req.Prompt, historyContext)
		sqlMsgs := []Message{
			{Role: "system", Content: "You are an expert SQL assistant that can create, alter, drop and query tables. Convert natural language into SQL using the schema below."},
			{Role: "user", Content: sqlPrompt},
		}
		llmSQL, err := connectLLM(sqlMsgs)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "LLM error during SQL generation: " + err.Error()})
			return
		}
		sqlQuery = strings.TrimSpace(llmSQL)
	} else {
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

		// build and call the SQL-generation LLM with history
		sqlPrompt := buildSQLPromptWithHistory(relevantSchema, req.Prompt, historyContext)
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
		sqlType := getOperationType(sqlQuery)
		c.JSON(http.StatusOK, gin.H{
			"needs_confirmation": true,
			"sql_preview":        sqlQuery,
			"message":            fmt.Sprintf("This %s may modify your database. Please confirm before execution.", sqlType),
			"sql_type":           sqlType,
			"session_id":         sessionID,
		})
		return
	}

	// 5) Execute and return result_table
	rows, err := db.Query(sqlQuery)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "SQL execution failed: " + err.Error(),
			"sql":   sqlQuery,
		})
		return
	}
	defer rows.Close()

	cols, _ := rows.Columns()
	result := []map[string]interface{}{}

	// Check if this is a modification query by looking at statement type
	sqlCmd := strings.TrimSpace(strings.ToUpper(sqlQuery))
	isModification := strings.HasPrefix(sqlCmd, "INSERT") ||
		strings.HasPrefix(sqlCmd, "UPDATE") ||
		strings.HasPrefix(sqlCmd, "DELETE") ||
		strings.HasPrefix(sqlCmd, "CREATE") ||
		strings.HasPrefix(sqlCmd, "DROP") ||
		strings.HasPrefix(sqlCmd, "ALTER")

	// Get affected rows info
	var affected int64 = 0
	if isModification {
		if ra, err := getAffectedRows(db); err == nil {
			affected = ra
		}
	}

	// Process result rows
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

	// Build response
	response := gin.H{
		"sql":          sqlQuery,
		"result_table": result,
		"session_id":   sessionID,
	}

	// Add sql_type and affected for modification queries
	if isModification {
		response["sql_type"] = getOperationType(sqlQuery)
		response["affected"] = affected
		response["message"] = fmt.Sprintf("%s completed. %d rows affected.", getOperationType(sqlQuery), affected)
	}

	// Add to conversation history
	responseText := fmt.Sprintf("Executed successfully. %d rows affected.", affected)
	if len(result) > 0 {
		responseText = fmt.Sprintf("Returned %d results.", len(result))
	}
	addToHistory(sessionID, req.Prompt, sqlQuery, responseText)

	c.JSON(http.StatusOK, response)
}

func getConversationHistory(sessionID, clientIP string) *ConversationHistory {
	conversationMutex.Lock()
	defer conversationMutex.Unlock()

	history, exists := conversations[sessionID]
	if !exists {
		history = &ConversationHistory{
			Items:    []HistoryItem{},
			ClientIP: clientIP,
			LastUsed: time.Now(),
		}
		conversations[sessionID] = history
	} else {
		history.LastUsed = time.Now()
	}

	return history
}

func addToHistory(sessionID, prompt, sql, response string) {
	conversationMutex.Lock()
	defer conversationMutex.Unlock()

	history, exists := conversations[sessionID]
	if !exists {
		return // Should not happen but handle just in case
	}

	// Add new item
	history.Items = append(history.Items, HistoryItem{
		Prompt:   prompt,
		SQL:      sql,
		Response: response,
		Time:     time.Now(),
	})

	// Trim if needed
	if len(history.Items) > MAX_HISTORY_ITEMS {
		history.Items = history.Items[len(history.Items)-MAX_HISTORY_ITEMS:]
	}

	history.LastUsed = time.Now()
}

func getAffectedRows(db *sql.DB) (int64, error) {
	var affected int64
	err := db.QueryRow("SELECT ROW_COUNT()").Scan(&affected)
	return affected, err
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

For CREATE TABLE operations, use appropriate data types and constraints.

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
If the user wants to create a new table or perform operations not related to specific existing tables, return "!!".
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

func buildSQLPromptWithHistory(schema map[string]TableInfo, userQuery, historyContext string) string {
	var sb strings.Builder

	for table, info := range schema {
		sb.WriteString(fmt.Sprintf("Table: %s\n", table))
		for _, col := range info.Columns {
			sb.WriteString(fmt.Sprintf("- %s: %s\n", col.Name, col.DataType))
		}
		sb.WriteString("\n")
	}

	return fmt.Sprintf(`
You are a SQL expert. Given the schema, conversation history, and user request, generate a valid SQL query.
Only return the SQL. Do not explain anything. Do not format the sql. Give it in raw text.

For CREATE TABLE operations, use appropriate data types and constraints.

### Schema
%s

### Conversation History
%s

### Request
%s

### SQL
`, sb.String(), historyContext, userQuery)
}
