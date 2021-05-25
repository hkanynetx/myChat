package main

import (
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"myChat/apiV1"
	"net/http"
	"os"
)

/* 主函数 */
func main() {

	// 设置日志
	gin.DisableConsoleColor()
	f, err := os.Create("./logs/run.log")
	if err != nil {
		log.Println("Could not open log.")
		panic(err)
	}
	gin.DefaultWriter = io.MultiWriter(f)

	gin.SetMode(gin.ReleaseMode)
	// 创建实例
	r := gin.Default()

	// ApiV1
	api := r.Group("/api/v1")
	ws := r.Group("/ws")

	api.GET("/length", apiV1.GetAllMembers)

	// WebSocket
	ws.GET("/connect", apiV1.Connect)

	// 错误处理
	r.NoRoute(func(context *gin.Context) {
		context.JSON(http.StatusNotFound, gin.H{"Status": 404, "msg": "Page Not Found"})
	})

	err = r.Run("localhost:10888")
	if err != nil {
		panic(err)
	}
}
