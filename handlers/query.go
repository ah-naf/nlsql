package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

var destructiveRE = regexp.MustCompile(`(?i)^\s*(INSERT|UPDATE|DELETE|DROP|ALTER|CREATE)`)

func needsConfirmation(sql string) bool {
	return destructiveRE.MatchString(strings.TrimSpace(sql))
}

func ShowQueryPage(c *gin.Context) {
	sess := sessions.Default(c)

	raw, ok := sess.Get("schema").(string)
	if !ok || raw == "" {
		c.String(http.StatusBadRequest, "no schema in session; please re‑select your database")
		return
	}

	var schema map[string][]string
	if err := json.Unmarshal([]byte(raw), &schema); err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("failed to parse schema from session: %v", err))
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

func HandleNLQuery(c *gin.Context) {
	var req struct {
		NLQuery   string `json:"nl_query"`
		Confirmed bool   `json:"confirmed"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	sess := sessions.Default(c)
	raw, _ := sess.Get("schema").(string)
	var schema map[string][]string
	if err := json.Unmarshal([]byte(raw), &schema); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "bad schema"})
		return
	}

	// sql := buildPrompt(schema, req.NLQuery)
	sqlP := "SELECT * FROM users"

	if needsConfirmation(sqlP) && !req.Confirmed {
		c.JSON(http.StatusOK, gin.H{
			"needs_confirmation": true,
			"sql_preview":        sqlP,
		})
		return
	}

	connStr, ok := sess.Get("connection_string").(string)
	if !ok || connStr == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no connection string in session"})
		return
	}
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("DB connect error: %v", err)})
		return
	}
	defer db.Close()

	upper := strings.ToUpper(strings.TrimSpace(sqlP))
	if strings.HasPrefix(upper, "SELECT") {
		rows, err := db.Query(sqlP)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("query error: %v", err)})
			return
		}
		defer rows.Close()

		cols, _ := rows.Columns()
		result := []map[string]interface{}{}

		for rows.Next() {
			// create a slice of interface{}'s to hold column values, and a second
			// slice to contain pointers to each item in the values slice.
			values := make([]interface{}, len(cols))
			pointers := make([]interface{}, len(cols))
			for i := range values {
				pointers[i] = &values[i]
			}

			if err := rows.Scan(pointers...); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("scan error: %v", err)})
				return
			}

			// build a map for this row, keyed by column name
			rowMap := make(map[string]interface{})
			for i, col := range cols {
				rowMap[col] = values[i]
			}
			result = append(result, rowMap)
		}

		c.JSON(http.StatusOK, gin.H{
			"status":      "ok",
			"sql_preview": sqlP,
			"table":       result,
		})
		return
	}

	// 7) non‑SELECT: execute and return a message
	res, err := db.Exec(sqlP)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("exec error: %v", err)})
		return
	}
	affected, _ := res.RowsAffected()

	c.JSON(http.StatusOK, gin.H{
		"status":      "ok",
		"sql_preview": sqlP,
		"message":     fmt.Sprintf("Query OK, %d rows affected", affected),
	})
}
