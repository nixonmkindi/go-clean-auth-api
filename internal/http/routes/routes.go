package routes

import (
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"github.com/yourusername/go-clean-auth-api/internal/http/controllers"
	custommw "github.com/yourusername/go-clean-auth-api/internal/http/middleware"
)

type Dependencies struct {
	Health             *controllers.HealthController
	Auth               *controllers.AuthController
	APIClient          *controllers.APIClientController
	Project            *controllers.ProjectController
	Task               *controllers.TaskController
	JWTSecret          string
	APIClientValidator custommw.APIClientValidator
}

func Register(e *echo.Echo, deps Dependencies) {
	e.Use(echomw.RequestID())
	e.Use(echomw.CORSWithConfig(echomw.CORSConfig{
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Authorization", "Content-Type", "X-API-Key", "X-API-Secret", "X-Request-ID"},
		AllowOrigins: []string{"*"},
	}))
	e.Use(echomw.Gzip())
	e.Use(echomw.Recover())
	e.Use(custommw.SecurityHeaders())

	e.GET("/health", deps.Health.Check)
	e.GET("/docs/openapi.yaml", echo.WrapHandler(httpFileServer("docs/openapi.yaml")))

	v1 := e.Group("/api/v1")
	auth := v1.Group("/auth")
	auth.POST("/register", deps.Auth.Register)
	auth.POST("/login", deps.Auth.Login)

	jwtProtected := v1.Group("", custommw.JWTAuth(deps.JWTSecret))
	jwtProtected.POST("/api-clients", deps.APIClient.Register)

	jwtProtected.POST("/projects", deps.Project.Create)
	jwtProtected.GET("/projects", deps.Project.List)
	jwtProtected.GET("/projects/:id", deps.Project.Get)
	jwtProtected.PUT("/projects/:id", deps.Project.Update)
	jwtProtected.DELETE("/projects/:id", deps.Project.Delete)

	jwtProtected.POST("/tasks", deps.Task.Create)
	jwtProtected.GET("/tasks", deps.Task.List)
	jwtProtected.GET("/tasks/:id", deps.Task.Get)
	jwtProtected.PUT("/tasks/:id", deps.Task.Update)
	jwtProtected.PATCH("/tasks/:id/assign", deps.Task.Assign)
	jwtProtected.PATCH("/tasks/:id/status", deps.Task.UpdateStatus)
	jwtProtected.DELETE("/tasks/:id", deps.Task.Delete)

	apiKeyProtected := v1.Group("", custommw.APIKeyAuth(deps.APIClientValidator))
	apiKeyProtected.GET("/client/projects/:project_id/tasks", deps.Task.ListByProjectForAPIClient)
}
