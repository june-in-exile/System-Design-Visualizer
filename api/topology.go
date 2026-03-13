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

	// 這裡的 CORS 設定與前端部署一致
	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// Vercel 自動對應到 /api/topology，但 Gin 也需要註冊這個路徑
	// 同時加上 Catch-all，確保無論 rewrite 怎麼變都能接到
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
