package middleware

import (
	"context"
	"net/http"

	"github.com/yourusername/go-clean-auth-api/internal/domain"
	"github.com/yourusername/go-clean-auth-api/internal/pkg/apperror"

	"github.com/labstack/echo/v4"
)

type APIClientValidator interface {
	ValidateAPIClient(ctx context.Context, key, secret string) (*domain.APIClient, error)
}

func APIKeyAuth(validator APIClientValidator) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			apiKey := c.Request().Header.Get(headerAPIKey)
			apiSecret := c.Request().Header.Get(headerAPISecret)
			if apiKey == "" || apiSecret == "" {
				return apperror.New(http.StatusUnauthorized, "MISSING_API_CREDENTIALS", "missing api key headers", nil, nil)
			}

			client, err := validator.ValidateAPIClient(c.Request().Context(), apiKey, apiSecret)
			if err != nil {
				return err
			}

			c.Set(ContextAPIClientIDKey, client.ID)
			c.Set(ContextAPIOwnerIDKey, client.OwnerID)
			return next(c)
		}
	}
}
