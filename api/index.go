package handler

import (
	"net/http"

	backendHandler "github.com/architectmind/backend/handler"
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

	// 定義路由。在 Vercel 中，這個 Handler 會被掛載在 /api/index
	// 我們透過別名 backendHandler 呼叫原本的 PostTopology
	app.POST("/api/topology", backendHandler.PostTopology)
	app.POST("/api/index/topology", backendHandler.PostTopology)
}

// Handler 是 Vercel Go Runtime 的進入點
func Handler(w http.ResponseWriter, r *http.Request) {
	app.ServeHTTP(w, r)
}
