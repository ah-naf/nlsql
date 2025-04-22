package main

import (
	"log"
	"nlsql/controller"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Routes
	r.POST("/connect", controller.ConnectDB)
	r.POST("/create", controller.CreateDB)
	r.POST("/delete", controller.DeleteDB)

	r.POST("/query", controller.HandleNLQuery)

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
