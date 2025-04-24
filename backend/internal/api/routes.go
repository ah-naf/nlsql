package api

import (
	"github.com/gin-gonic/gin"
)

// SetupRouter registers all routes and returns a *gin.Engine.
func SetupRouter() *gin.Engine {
	r := gin.Default()

	// NL→SQL
	r.POST("/query", HandleNLQuery)

	// DB management
	r.GET("/databases", GetDatabases)
	r.POST("/create", CreateDB)
	r.POST("/delete", DeleteDB)
	r.POST("/connect", ConnectDB)

	// Schema introspection
	r.GET("/schema", GetSchema)
	r.GET("/schema/:tableName", GetTableSchema)

	return r
}
