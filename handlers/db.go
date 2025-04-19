package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"nlsql/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

func ShowConnectForm(c *gin.Context) {
	sess := sessions.Default(c)
	c.HTML(http.StatusOK, "connect.html", gin.H{
		"Host":     sess.Get("host"),
		"Port":     sess.Get("port"),
		"User":     sess.Get("user"),
		"DBName":   sess.Get("dbname"),
		"ErrorMsg": sess.Get("conn_error"),
	})
	sess.Delete("conn_error")
	sess.Save()
}

func CreateDB(c *gin.Context) {
	sess := sessions.Default(c)
	host := sess.Get("host").(string)
	port := sess.Get("port").(string)
	user := sess.Get("user").(string)
	pass := sess.Get("pass").(string)

	newDB := c.PostForm("newdb")
	if newDB == "" {
		sess.Set("create_error", "Database name cannot be blank")
		sess.Save()
		c.Redirect(http.StatusSeeOther, "/select")
		return
	}

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable",
		host, port, user, pass,
	)

	db, err := sql.Open("postgres", connStr)
	if err == nil {
		err = db.Ping()
	}
	if err != nil {
		sess.Set("create_error", fmt.Sprintf("Connection error: %v", err))
		sess.Save()
		c.Redirect(http.StatusSeeOther, "/select")
		return
	}
	defer db.Close()

	if _, err := db.Exec(fmt.Sprintf("CREATE DATABASE %s;", pq.QuoteIdentifier(newDB))); err != nil {
		sess.Set("create_error", fmt.Sprintf("Could not create database: %v", err))
		sess.Save()
		c.Redirect(http.StatusSeeOther, "/select")
		return
	}

	selConn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, pass, newDB,
	)
	sess.Set("connection_string", selConn)
	sess.Save()

	c.Redirect(http.StatusSeeOther, "/query")
}

func ConnectDB(c *gin.Context) {
	host := c.PostForm("host")
	port := c.PostForm("port")
	user := c.PostForm("user")
	pass := c.PostForm("pass")
	dbname := c.PostForm("dbname")

	if dbname == "" {
		dbname = "postgres"
	}

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, pass, dbname,
	)
	db, err := sql.Open("postgres", connStr)
	if err == nil {
		err = db.Ping()
	}
	if err != nil {
		sess := sessions.Default(c)
		sess.Set("conn_error", err.Error())
		sess.Save()
		c.Redirect(http.StatusSeeOther, "/")
		return
	}
	db.Close()

	// Save connection details
	sess := sessions.Default(c)
	sess.Set("host", host)
	sess.Set("port", port)
	sess.Set("user", user)
	sess.Set("pass", pass)
	sess.Set("dbname", dbname)
	sess.Save()

	c.Redirect(http.StatusSeeOther, "/select")
}

// ShowDBForm lists available databases using saved session details
func ShowDBForm(c *gin.Context) {
	sess := sessions.Default(c)
	host := sess.Get("host").(string)
	port := sess.Get("port").(string)
	user := sess.Get("user").(string)
	pass := sess.Get("pass").(string)
	dbname := sess.Get("dbname").(string)

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, pass, dbname,
	)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("DB connection error: %v", err))
		return
	}
	defer db.Close()

	dbs, err := models.GetDatabases(db)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error fetching databases: %v", err))
		return
	}

	c.HTML(http.StatusOK, "select.html", gin.H{
		"Databases": dbs,
	})
}

// SelectDB handles database selection and shows schema details
func SelectDB(c *gin.Context) {
	sess := sessions.Default(c)
	host := sess.Get("host").(string)
	port := sess.Get("port").(string)
	user := sess.Get("user").(string)
	pass := sess.Get("pass").(string)

	selDB := c.PostForm("db")
	// Connect to selected DB
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, pass, selDB,
	)
	sess.Set("connection_string", connStr)
	sess.Save()

	c.Redirect(http.StatusSeeOther, "/query")
}

// ResetSession clears all saved connection data
func ResetSession(c *gin.Context) {
	sess := sessions.Default(c)
	sess.Clear()
	sess.Save()
	c.Redirect(http.StatusSeeOther, "/")
}
