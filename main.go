package main

import (
	"log"
	"nlsql/handlers"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Session middleware
	store := cookie.NewStore([]byte("secret-key"))
	r.Use(sessions.Sessions("nlsql_session", store))

	// Load templates
	r.LoadHTMLGlob("templates/*.html")
	r.Static("/static", "./static")

	// Routes
	r.GET("/", handlers.ShowConnectForm)
	r.POST("/connect", handlers.ConnectDB)
	r.GET("/select", handlers.ShowDBForm)
	r.POST("/select", handlers.SelectDB)
	r.GET("/reset", handlers.ResetSession)

	r.GET("/query", handlers.ShowQueryPage)
	r.POST("/query", handlers.HandleNLQuery)

	r.POST("/createdb", handlers.CreateDB)

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
