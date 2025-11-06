package gateway

import (
	"fmt"
	"net/http/httputil"
	"os"

	"github.com/PegasusMKD/svedprint-go/internal/gateway/db/sqlc"
	requestlogger "github.com/PegasusMKD/svedprint-go/internal/gateway/request_logger"
	"github.com/PegasusMKD/svedprint-go/pkg/config"
	"github.com/PegasusMKD/svedprint-go/pkg/database"
	"github.com/gin-gonic/gin"
)

type GinServer struct {
	addr   string
	engine *gin.Engine

	requestLogWriter *requestlogger.LogWriter

	proxies map[string]*httputil.ReverseProxy
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

	server := &GinServer{engine: router, addr: addr, proxies: make(map[string]*httputil.ReverseProxy)}
	server.setupProxyAndAuth()

	return server
}

func setupSqlc(cfg *config.Config) *sqlc.Queries {
	dbConfig := database.GetConfig(cfg.DatabaseURL, cfg.DatabaseMaxConns, cfg.DatabaseMaxIdleConns, cfg.DatabaseConnLifetime)
	migrationPath := fmt.Sprintf("db/%s/migrations", cfg.ServiceName)
	database.RunMigrations(dbConfig.URL, migrationPath)
	return sqlc.New(database.SetupDatabasePool(dbConfig))
}

func (gs *GinServer) setupProxyAndAuth() {
	configureProxies(gs)
	gs.engine.Any("/api/svedprint/*path", gs.ProxyTo("svedprint"))
	gs.engine.Any("/api/admin/*path", gs.ProxyTo("admin"))
	gs.engine.Any("/api/print/*path", gs.ProxyTo("print"))
}

func setupMiddleware(router *gin.Engine) {
	router.Use(gin.Logger())
}

func setupRoutes(router *gin.Engine) {
	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"status": "healthy"})
	})
}
