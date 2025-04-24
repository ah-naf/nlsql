package api

import (
	"net/http"

	"nlsql/internal/db"
	"nlsql/internal/models"

	"github.com/gin-gonic/gin"
)

// GetSchema handles GET /schema
func GetSchema(c *gin.Context) {
	var req models.DBRequest
	req.Host = c.Query("host")
	req.Port = c.Query("port")
	req.User = c.Query("user")
	req.Pass = c.Query("pass")
	req.DBName = c.Query("dbname")

	if req.Host == "" || req.Port == "" || req.User == "" || req.Pass == "" || req.DBName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required params"})
		return
	}
	conn, err := db.OpenConnection(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer conn.Close()

	if c.Query("brief") == "true" {
		list, err := db.BriefSchema(conn)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, models.BriefSchemaResponse{Tables: list})
		return
	}

	full, err := db.LoadFullSchema(conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.SchemaResponse{Schema: full})
}

// GetTableSchema handles GET /schema/:tableName
func GetTableSchema(c *gin.Context) {
	var req models.DBRequest
	req.Host = c.Query("host")
	req.Port = c.Query("port")
	req.User = c.Query("user")
	req.Pass = c.Query("pass")
	req.DBName = c.Query("dbname")
	table := c.Param("tableName")

	if req.Host == "" || req.Port == "" || req.User == "" || req.Pass == "" || req.DBName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required params"})
		return
	}
	conn, err := db.OpenConnection(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer conn.Close()

	full, err := db.LoadFullSchema(conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ti, ok := full[table]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "table not found"})
		return
	}
	c.JSON(http.StatusOK, models.TableSchemaResponse{Table: ti})
}
