package main

import (
	"net/http"

	"github.com/architectmind/backend/handler"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var app *gin.Engine

func init() {
	gin.SetMode(gin.ReleaseMode)
	app = gin.New()
	app.Use(gin.Recovery())

	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Vercel 部署環境下前後端同網域，或是可依需求限縮
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// 定義路由。在 Vercel 中，我們可以直接將 API 邏輯掛載在根路徑或加上 /api 前綴
	// 由於這個檔案被命名為 topology.go，請求 /api/topology 會直接調用此 Handler
	app.POST("/api/topology", handler.PostTopology)
	app.POST("/topology", handler.PostTopology) // 為了保險也支援不帶前綴的路徑
}

// Handler 是 Vercel Go Runtime 的進入點
func Handler(w http.ResponseWriter, r *http.Request) {
	app.ServeHTTP(w, r)
}
