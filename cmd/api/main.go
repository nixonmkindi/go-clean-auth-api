package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/yourusername/go-clean-auth-api/internal/config"
	"github.com/yourusername/go-clean-auth-api/internal/database"
	"github.com/yourusername/go-clean-auth-api/internal/http/controllers"
	custommw "github.com/yourusername/go-clean-auth-api/internal/http/middleware"
	"github.com/yourusername/go-clean-auth-api/internal/http/responses"
	"github.com/yourusername/go-clean-auth-api/internal/http/routes"
	"github.com/yourusername/go-clean-auth-api/internal/pkg/logger"
	customvalidator "github.com/yourusername/go-clean-auth-api/internal/pkg/validator"
	"github.com/yourusername/go-clean-auth-api/internal/repository/postgres"
	"github.com/yourusername/go-clean-auth-api/internal/usecase"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load error: %v", err)
	}

	appLogger := logger.New(cfg.LogLevel)
	ctx := context.Background()

	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		appLogger.Error("database_connection_error", "error", err.Error())
		os.Exit(1)
	}
	defer pool.Close()

	userRepo := postgres.NewUserRepository(pool)
	apiClientRepo := postgres.NewAPIClientRepository(pool)
	projectRepo := postgres.NewProjectRepository(pool)
	taskRepo := postgres.NewTaskRepository(pool)

	authUsecase := usecase.NewAuthUsecase(userRepo, apiClientRepo, cfg.JWTSecret, cfg.JWTTTLMinutes)
	apiClientUsecase := usecase.NewAPIClientUsecase(apiClientRepo)
	projectUsecase := usecase.NewProjectUsecase(projectRepo, cfg.DefaultPageSize, cfg.MaxPageSize)
	taskUsecase := usecase.NewTaskUsecase(taskRepo, projectRepo, userRepo, cfg.DefaultPageSize, cfg.MaxPageSize)

	healthController := controllers.NewHealthController()
	authController := controllers.NewAuthController(authUsecase)
	apiClientController := controllers.NewAPIClientController(apiClientUsecase)
	projectController := controllers.NewProjectController(projectUsecase)
	taskController := controllers.NewTaskController(taskUsecase)

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Validator = customvalidator.New()
	e.HTTPErrorHandler = responses.HTTPErrorHandler(appLogger)
	e.Use(custommw.RequestLogger(appLogger))

	routes.Register(e, routes.Dependencies{
		Health:             healthController,
		Auth:               authController,
		APIClient:          apiClientController,
		Project:            projectController,
		Task:               taskController,
		JWTSecret:          cfg.JWTSecret,
		APIClientValidator: authUsecase,
	})

	serverErr := make(chan error, 1)
	go func() {
		addr := ":" + cfg.AppPort
		appLogger.Info("server_starting", "address", addr)
		serverErr <- e.Start(addr)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			appLogger.Error("server_error", "error", err.Error())
			os.Exit(1)
		}
	case <-quit:
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := e.Shutdown(shutdownCtx); err != nil {
			appLogger.Error("server_shutdown_error", "error", err.Error())
			os.Exit(1)
		}
	}
}
