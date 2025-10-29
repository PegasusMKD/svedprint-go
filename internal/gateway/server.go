package gateway

import (
	"fmt"
	"os"

	"github.com/PegasusMKD/svedprint-go/internal/gateway/db/sqlc"
	"github.com/PegasusMKD/svedprint-go/pkg/config"
	"github.com/PegasusMKD/svedprint-go/pkg/database"
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

	cfg, err := config.Load("gateway")
	if err != nil {
		panic("Failed loading config for gateway!")
	}

	setupSqlc(cfg)

	router := gin.Default()

	setupMiddleware(router)
	setupRoutes(router)

	return &GinServer{engine: router, addr: addr}
}

func setupSqlc(cfg *config.Config) *sqlc.Queries {
	dbConfig := database.GetConfig(cfg.DatabaseURL, cfg.DatabaseMaxConns, cfg.DatabaseMaxIdleConns, cfg.DatabaseConnLifetime)
	migrationPath := fmt.Sprintf("db/%s/migrations", cfg.ServiceName)
	database.RunMigrations(dbConfig.URL, migrationPath)
	return sqlc.New(database.SetupDatabasePool(dbConfig))
}

func setupMiddleware(router *gin.Engine) {
	router.Use(gin.Logger())
}

func setupRoutes(router *gin.Engine) {
	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"status": "healthy"})
	})
}
