package svedprintadmin

import (
	"fmt"
	"log"
	"os"

	"github.com/PegasusMKD/svedprint-go/internal/svedprint-admin/db/sqlc"
	"github.com/PegasusMKD/svedprint-go/internal/svedprint-admin/handlers"
	"github.com/PegasusMKD/svedprint-go/internal/svedprint-admin/repositories"
	"github.com/PegasusMKD/svedprint-go/internal/svedprint-admin/services"
	"github.com/PegasusMKD/svedprint-go/pkg/config"
	"github.com/PegasusMKD/svedprint-go/pkg/database"
	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/user"
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

	cfg, err := config.Load("svedprint-admin")
	if err != nil {
		panic("Failed loading config for svedprint!")
	}

	queries := setupSqlc(cfg)

	router := gin.Default()

	setupMiddleware(router)
	setupRoutes(router, queries)

	return &GinServer{engine: router, addr: addr}
}

func setupSqlc(cfg *config.Config) *sqlc.Queries {
	dbConfig := database.GetConfig(cfg.DatabaseURL, cfg.DatabaseMaxConns, cfg.DatabaseMaxIdleConns, cfg.DatabaseConnLifetime)
	migrationPath := fmt.Sprintf("db/%s/migrations", cfg.ServiceName)
	if err := database.RunMigrations(dbConfig.URL, migrationPath); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	return sqlc.New(database.SetupDatabasePool(dbConfig))
}

func configureClerkUserClient() *user.Client {
	clerkConfig := &clerk.ClientConfig{}
	clerkSecretKey := os.Getenv("CLERK_SECRET_KEY")
	if clerkSecretKey == "" {
		log.Fatal("CLERK_SECRET_KEY is required")
	}
	clerkConfig.Key = &clerkSecretKey
	return user.NewClient(clerkConfig)
}

func setupMiddleware(router *gin.Engine) {
	router.Use(gin.Logger())
}

func setupRoutes(router *gin.Engine, queries *sqlc.Queries) {
	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"status": "healthy"})
	})

	clerkUserClient := configureClerkUserClient()

	teacherRepository := repositories.NewTeacherRepository(queries)
	teacherService := services.NewTeacherService(teacherRepository, clerkUserClient)
	registrationHandler := handlers.NewRegistrationHandler(teacherService)

	router.POST("/register", registrationHandler.RegisterUser)
}
