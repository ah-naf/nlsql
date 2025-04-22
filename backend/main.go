package main

import (
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"nlsql/handlers"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

//go:embed static/favicon.svg
//go:embed static/js/*
//go:embed static/css/*
//go:embed templates/*.html
var embeddedFiles embed.FS

func main() {
	staticFS, err := fs.Sub(embeddedFiles, "static")
	if err != nil {
		log.Fatal("failed to create static sub‑fs:", err)
	}

	r := gin.Default()

	// Session middleware
	store := cookie.NewStore([]byte("secret-key"))
	r.Use(sessions.Sessions("nlsql_session", store))

	tmpl := template.Must(template.ParseFS(embeddedFiles, "templates/*.html"))
	r.SetHTMLTemplate(tmpl)

	r.GET("/favicon.svg", func(c *gin.Context) {
		// special route for favicon
		data, _ := embeddedFiles.ReadFile("static/favicon.svg")
		c.Header("Content-Type", "image/svg+xml")
		c.Writer.Write(data)
	})
	r.StaticFS("/static", http.FS(staticFS))

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
