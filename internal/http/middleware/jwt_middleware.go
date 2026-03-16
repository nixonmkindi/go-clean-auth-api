package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/nixonmkindi/go-clean-auth-api/internal/pkg/apperror"
)

func JWTAuth(jwtSecret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				return apperror.New(http.StatusUnauthorized, "MISSING_BEARER_TOKEN", "missing or invalid bearer token", nil, nil)
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method")
				}
				return []byte(jwtSecret), nil
			})
			if err != nil || !token.Valid {
				return apperror.New(http.StatusUnauthorized, "INVALID_TOKEN", "invalid or expired token", nil, err)
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return apperror.New(http.StatusUnauthorized, "INVALID_TOKEN_CLAIMS", "invalid token claims", nil, nil)
			}

			subject, ok := claims["sub"].(string)
			if !ok || subject == "" {
				return apperror.New(http.StatusUnauthorized, "INVALID_TOKEN_SUBJECT", "invalid token subject", nil, nil)
			}
			userID, err := strconv.ParseInt(subject, 10, 64)
			if err != nil {
				return apperror.New(http.StatusUnauthorized, "INVALID_TOKEN_SUBJECT", "invalid token subject", nil, err)
			}

			c.Set(ContextUserIDKey, userID)
			return next(c)
		}
	}
}
