package svedprintprint

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
)

type GinServer struct {
	addr   string
	engine *gin.Engine
}

func (gs *GinServer) Run() {
	gs.engine.Run(gs.addr)
}

func NewServer() *GinServer {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	addr := fmt.Sprintf(":%s", port)

	router := gin.Default()

	setupMiddleware(router)
	setupRoutes(router)

	return &GinServer{engine: router, addr: addr}
}

func setupMiddleware(router *gin.Engine) {
	router.Use(gin.Logger())
}

func setupRoutes(router *gin.Engine) {
	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"status": "healthy"})
	})
}
