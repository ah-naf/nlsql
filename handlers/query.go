package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

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
