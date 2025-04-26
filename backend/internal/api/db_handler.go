// api/db_handler.go
package api

import (
	"net/http"

	"nlsql/internal/db"
	"nlsql/internal/models"

	"github.com/gin-gonic/gin"
)

// GetDatabases handles GET /databases
func GetDatabases(c *gin.Context) {
	var req models.DBRequest
    req.Host = c.Query("host")
    req.Port = c.Query("port")
    req.User = c.Query("user")
    req.Pass = c.Query("pass")
    req.DBName = c.Query("dbname")
    req.Provider = c.Query("provider")
    req.SSLMode = c.Query("sslmode")
    req.ConnectionString = c.Query("connectionString")

	if req.ConnectionString == "" && (req.Host == "" || req.Port == "" || req.User == "" || req.Pass == "") {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Either connection string or required params must be provided"})
        return
    }

	conn, err := db.OpenConnection(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer conn.Close()

	list, err := db.GetDatabases(conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.DatabaseListResponse{Databases: list})
}

// CreateDB handles POST /database/create
func CreateDB(c *gin.Context) {
	var req models.DBRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.DBName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "DBName cannot be blank"})
		return
	}
	conn, err := db.OpenAdminConnection(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer conn.Close()

	if err := db.CreateDatabase(conn, req.DBName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Database created", "dbname": req.DBName})
}

// DeleteDB handles POST /database/delete
func DeleteDB(c *gin.Context) {
	var req models.DBRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.DBName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "DBName cannot be blank"})
		return
	}
	conn, err := db.OpenAdminConnection(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer conn.Close()

	if err := db.DeleteDatabase(conn, req.DBName); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Database deleted", "dbname": req.DBName})
}

// ConnectDB handles POST /database/connect
func ConnectDB(c *gin.Context) {
	var req models.DBRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.ConnectionString == "" && (req.Host == "" || req.Port == "" || req.User == "" || req.Pass == "") {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Either connection string or required params must be provided"})
        return
    }

	conn, err := db.OpenConnection(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer conn.Close()

	if err := conn.Ping(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	list, err := db.GetDatabases(conn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"connection": req,
		"databases":  list,
	})
}
