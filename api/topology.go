package handler

import (
	"net/http"

	"github.com/architectmind/backend/logic"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var app *gin.Engine

func init() {
	gin.SetMode(gin.ReleaseMode)
	app = gin.New()
	app.Use(gin.Recovery())

	// CORS configuration consistent with frontend deployment
	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// Vercel automatically maps to /api/topology, but Gin also needs to register this path.
	// A catch-all is added to ensure it works regardless of rewrite rules.
	app.POST("/api/topology", logic.PostTopology)
	app.NoRoute(func(c *gin.Context) {
		if c.Request.Method == http.MethodPost {
			logic.PostTopology(c)
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	})
}

// Handler is the Vercel Go Runtime entry point.
func Handler(w http.ResponseWriter, r *http.Request) {
	app.ServeHTTP(w, r)
}
