package controllers

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/yourusername/go-clean-auth-api/internal/http/responses"
)

type HealthController struct{}

func NewHealthController() *HealthController {
	return &HealthController{}
}

func (h *HealthController) Check(c echo.Context) error {
	return responses.Success(c, http.StatusOK, map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
