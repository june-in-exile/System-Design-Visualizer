package main

import (
	"log"

	"github.com/architectmind/backend/handler"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	api := r.Group("/api")
	{
		api.POST("/topology", handler.PostTopology)
	}

	log.Println("ArchitectMind backend starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
