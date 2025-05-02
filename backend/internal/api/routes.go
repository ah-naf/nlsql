package api

import (
	"net/http"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SetupRouter registers all routes and returns a *gin.Engine.
func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // or specify: []string{"http://localhost:3000"}
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))

	// NL→SQL
	r.POST("/query", HandleNLQuery)
	r.POST("/execute-sql", HandleDirectSQL)

	// DB management
	r.GET("/databases", GetDatabases)
	r.POST("/create", CreateDB)
	r.POST("/delete", DeleteDB)
	r.POST("/connect", ConnectDB)

	// Schema introspection
	r.GET("/schema", GetSchema)
	r.GET("/schema/:tableName", GetTableSchema)

	r.Static("/assets", "../frontend/dist/assets")

	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		if c.Request.Method == "GET" && (path == "/" || path == "/select" || path == "/query") {
			c.File("../frontend/dist/index.html")
			return
		}

		if strings.Contains(path, ".") {
			c.Status(http.StatusNotFound)
			return
		}

		if c.Request.Method == "GET" {
			c.File("../frontend/dist/index.html")
			return
		}

		c.JSON(http.StatusNotFound, gin.H{"error": "API endpoint not found"})
	})

	return r
}
