package app

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
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

type Application struct {
	Config *config.Config
	Logger *slog.Logger
	Echo   *echo.Echo
	DB     *pgxpool.Pool
}

func New(ctx context.Context) (*Application, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	appLogger := logger.New(cfg.LogLevel)
	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Validator = customvalidator.New()
	e.HTTPErrorHandler = responses.HTTPErrorHandler(appLogger)
	e.Use(custommw.RequestLogger(appLogger))

	registerRoutes(e, cfg, pool)

	return &Application{
		Config: cfg,
		Logger: appLogger,
		Echo:   e,
		DB:     pool,
	}, nil
}

func (a *Application) Shutdown(ctx context.Context) error {
	if err := a.Echo.Shutdown(ctx); err != nil {
		return err
	}
	a.DB.Close()
	return nil
}

func registerRoutes(e *echo.Echo, cfg *config.Config, pool *pgxpool.Pool) {
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

	routes.Register(e, routes.Dependencies{
		Health:             healthController,
		Auth:               authController,
		APIClient:          apiClientController,
		Project:            projectController,
		Task:               taskController,
		JWTSecret:          cfg.JWTSecret,
		APIClientValidator: authUsecase,
	})
}
