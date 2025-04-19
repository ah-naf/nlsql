package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"nlsql/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
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

func ConnectDB(c *gin.Context) {
	host := c.PostForm("host")
	port := c.PostForm("port")
	user := c.PostForm("user")
	pass := c.PostForm("pass")
	dbname := c.PostForm("dbname")

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
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("DB connect error: %v", err))
		return
	}
	defer db.Close()

	schema, err := models.GetSchema(db)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Schema error: %v", err))
		return
	}

	b, err := json.Marshal(schema)
	if err != nil {
		c.String(http.StatusInternalServerError, "schema marshal error: %v", err)
		return
	}
	sess.Set("schema", string(b))
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
