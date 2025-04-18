package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func ShowQueryPage(c *gin.Context) {
	sess := sessions.Default(c)

	raw := sess.Get("schema")
	schema, ok := raw.(map[string][]string)
	if !ok {
		schema = make(map[string][]string)
	}

	fmt.Printf("▶ Loaded schema: %#v\n", schema)

	c.HTML(http.StatusOK, "query.html", gin.H{
		"Schema": schema,
	})
}
