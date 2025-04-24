package api

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"nlsql/internal/db"
	"nlsql/internal/llm"
	"nlsql/internal/models"
	"nlsql/pkg/utils"

	"github.com/gin-gonic/gin"
)

const (
	maxHistoryItems = 10
	historyExpiry   = 30 * time.Minute
)

var (
	conversations     = make(map[string]*models.ConversationHistory)
	conversationMutex sync.Mutex
)

func init() {
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			cleanupExpired()
		}
	}()
}

func cleanupExpired() {
	conversationMutex.Lock()
	defer conversationMutex.Unlock()
	now := time.Now()
	for id, hist := range conversations {
		if now.Sub(hist.LastUsed) > historyExpiry {
			delete(conversations, id)
		}
	}
}

func getHistory(sessionID, clientIP string) *models.ConversationHistory {
	conversationMutex.Lock()
	defer conversationMutex.Unlock()
	if h, ok := conversations[sessionID]; ok {
		h.LastUsed = time.Now()
		return h
	}
	h := &models.ConversationHistory{
		Items:    []models.HistoryItem{},
		ClientIP: clientIP,
		LastUsed: time.Now(),
	}
	conversations[sessionID] = h
	return h
}

func addToHistory(sessionID, prompt, sqlQ, responseText string) {
	conversationMutex.Lock()
	defer conversationMutex.Unlock()
	h, ok := conversations[sessionID]
	if !ok {
		return
	}
	h.Items = append(h.Items, models.HistoryItem{
		Prompt:   prompt,
		SQL:      sqlQ,
		Response: responseText,
		Time:     time.Now(),
	})
	if len(h.Items) > maxHistoryItems {
		h.Items = h.Items[len(h.Items)-maxHistoryItems:]
	}
	h.LastUsed = time.Now()
}

// HandleNLQuery is the POST /nlq handler.
func HandleNLQuery(c *gin.Context) {
	var req models.NLQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
		return
	}

	// Session
	sessionID := req.SessionID
	if sessionID == "" {
		sessionID = fmt.Sprintf("%s-%s", c.ClientIP(), req.Config.DBName)
	}
	history := getHistory(sessionID, c.ClientIP())

	// DB connection
	conn, err := db.OpenConnection(req.Config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB connection: " + err.Error()})
		return
	}
	defer conn.Close()

	// 1) Table detection
	isLikely := utils.IsDBOperation(req.Prompt)
	tables, err := db.GetTableNameList(conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Table fetch: " + err.Error()})
		return
	}
	detPrompt := llm.BuildTableDetectionPrompt(tables, req.Prompt)
	detResp, err := llm.Connect([]models.Message{
		{Role: "system", Content: "Pick relevant table names or !!."},
		{Role: "user", Content: detPrompt},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "LLM detect: " + err.Error()})
		return
	}
	if strings.TrimSpace(detResp) == "!!" && !isLikely {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Could not find relevant tables for your query. Please rephrase with specific table references.",
			"session_id": sessionID,
		})
		return
	}

	// 2) Build history context
	histCtx := ""
	if len(history.Items) > 0 {
		histCtx = "Previous:\n"
		for i, item := range history.Items {
			histCtx += fmt.Sprintf("User: %s\nSQL: %s\n\n", item.Prompt, item.SQL)
			if i >= 4 {
				break
			}
		}
	}

	// 3) NL→SQL
	var sqlQuery string
	if isLikely && (strings.TrimSpace(detResp) == "!!" || strings.Contains(strings.ToLower(req.Prompt), "table")) {
		fullSchema, err := db.LoadFullSchema(conn)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Schema load: " + err.Error()})
			return
		}
		prompt := llm.BuildSQLPromptWithHistory(fullSchema, req.Prompt, histCtx)
		out, err := llm.Connect([]models.Message{
			{Role: "system", Content: "Expert SQL assistant for DDL & queries."},
			{Role: "user", Content: prompt},
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "LLM SQL gen: " + err.Error()})
			return
		}
		sqlQuery = strings.TrimSpace(out)
	} else {
		detected := utils.ParseCSV(detResp)
		fullSchema, err := db.LoadFullSchema(conn)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Schema load: " + err.Error()})
			return
		}
		// filter to detected tables
		rel := make(map[string]models.TableInfo)
		for _, t := range detected {
			if info, ok := fullSchema[t]; ok {
				rel[t] = info
			}
		}
		prompt := llm.BuildSQLPromptWithHistory(rel, req.Prompt, histCtx)
		out, err := llm.Connect([]models.Message{
			{Role: "system", Content: "Expert SQL assistant for queries."},
			{Role: "user", Content: prompt},
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "LLM SQL gen: " + err.Error()})
			return
		}
		sqlQuery = strings.TrimSpace(out)
	}

	// 4) Confirm destructive
	if utils.NeedsConfirmation(sqlQuery) && !req.Confirmed {
		opType := utils.OperationType(sqlQuery)
		c.JSON(http.StatusOK, gin.H{
			"needs_confirmation": true,
			"sql_preview":        sqlQuery,
			"message":            fmt.Sprintf("This %s may modify your DB. Confirm to proceed.", opType),
			"sql_type":           opType,
			"session_id":         sessionID,
		})
		return
	}

	// 5) Execute
	isMod := utils.NeedsConfirmation(sqlQuery)
	var results []map[string]interface{}
	var affected int64

	if isMod {
		affected, err = db.ExecuteModification(conn, sqlQuery)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Exec failed: " + err.Error(), "sql": sqlQuery})
			return
		}
	} else {
		rows, err := conn.Query(sqlQuery)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Query failed: " + err.Error(), "sql": sqlQuery})
			return
		}
		defer rows.Close()
		cols, _ := rows.Columns()
		for rows.Next() {
			vals := make([]interface{}, len(cols))
			ptrs := make([]interface{}, len(cols))
			for i := range ptrs {
				ptrs[i] = &vals[i]
			}
			if err := rows.Scan(ptrs...); err != nil {
				continue
			}
			row := map[string]interface{}{}
			for i, n := range cols {
				row[n] = vals[i]
			}
			results = append(results, row)
		}
	}

	// 6) Build response
	resp := models.NLQueryResponse{
		SQL:       sqlQuery,
		SessionID: sessionID,
	}
	if isMod {
		resp.SQLType = utils.OperationType(sqlQuery)
		resp.Affected = affected
		resp.Message = fmt.Sprintf("%s done. %d rows affected.", resp.SQLType, affected)
	} else {
		resp.ResultTable = results
	}

	// 7) Save history
	msgText := fmt.Sprintf("Returned %d rows.", len(results))
	if isMod {
		msgText = fmt.Sprintf("%d rows affected.", affected)
	}
	addToHistory(sessionID, req.Prompt, sqlQuery, msgText)

	c.JSON(http.StatusOK, resp)
}
