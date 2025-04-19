package handlers

import (
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

	sql := buildPrompt(schema, req.NLQuery)

	if needsConfirmation(sql) && !req.Confirmed {
		c.JSON(http.StatusOK, gin.H{
			"needs_confirmation": true,
			"sql_preview":        sql,
		})
		return
	}
}
