package main

import (
	"log"
	"nlsql/controller"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Routes
	r.POST("/connect", controller.ConnectDB)
	r.POST("/create", controller.CreateDB)
	r.POST("/delete", controller.DeleteDB)

	r.POST("/query", controller.HandleNLQuery)

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
