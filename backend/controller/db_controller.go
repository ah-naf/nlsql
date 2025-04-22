package controller

import (
	"database/sql"
	"fmt"
	"net/http"
	"nlsql/models"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type DBRequest struct {
	Host   string `json:"host"`
	Port   int    `json:"port"`
	User   string `json:"user"`
	Pass   string `json:"pass"`
	DBName string `json:"dbname"`
}

func DeleteDB(c *gin.Context) {
	var req DBRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
		return
	}

	if req.DBName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Database name cannot be blank"})
		return
	}

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable",
		req.Host, req.Port, req.User, req.Pass,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Connection error: " + err.Error()})
		return
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ping failed: " + err.Error()})
		return
	}

	_, err = db.Exec(fmt.Sprintf("DROP DATABASE %s;", pq.QuoteIdentifier(req.DBName)))
	if err != nil {
		if strings.Contains(err.Error(), "being accessed by other users") {
			c.JSON(http.StatusConflict, gin.H{
				"error": fmt.Sprintf("Database %s is currently in use and cannot be dropped", req.DBName),
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Could not drop database: %s", err.Error()),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Database %s deleted successfully", req.DBName),
	})
}

func CreateDB(c *gin.Context) {
	var req DBRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
		return
	}

	if req.DBName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Database name cannot be blank"})
		return
	}

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable",
		req.Host, req.Port, req.User, req.Pass,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Connection error: " + err.Error()})
		return
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ping failed: " + err.Error()})
		return
	}

	createStmt := fmt.Sprintf("CREATE DATABASE %s;", pq.QuoteIdentifier(req.DBName))
	if _, err := db.Exec(createStmt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create database: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    fmt.Sprintf("Database %s created successfully", req.DBName),
		"connection": req,
	})
}

func ConnectDB(c *gin.Context) {
	var req DBRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
		return
	}

	if req.DBName == "" {
		req.DBName = "postgres"
	}

	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		req.Host, req.Port, req.User, req.Pass, req.DBName,
	)
	fmt.Println("DB Conn:", connStr)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Connection failed: " + err.Error()})
		return
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ping failed: " + err.Error()})
		return
	}

	dbs, err := models.GetDatabases(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get databases: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"databases": dbs,
		"connection": DBRequest{
			Host:   req.Host,
			Port:   req.Port,
			User:   req.User,
			Pass:   req.Pass,
			DBName: req.DBName,
		},
	})
}
