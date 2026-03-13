package handler

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var app *gin.Engine

func init() {
	gin.SetMode(gin.ReleaseMode)
	app = gin.New()
	app.Use(gin.Recovery())

	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	app.POST("/api/topology", PostTopology)
	app.POST("/api/handler/topology", PostTopology)
}

// Handler is the Vercel Go Runtime entry point.
func Handler(w http.ResponseWriter, r *http.Request) {
	app.ServeHTTP(w, r)
}
