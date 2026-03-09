package middleware

import (
	"log/slog"
	"time"

	"github.com/labstack/echo/v4"
)

func RequestLogger(log *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)
			latency := time.Since(start)

			status := c.Response().Status
			if status == 0 {
				status = 200
			}

			log.Info("http_request",
				"method", c.Request().Method,
				"path", c.Path(),
				"uri", c.Request().RequestURI,
				"status", status,
				"latency_ms", latency.Milliseconds(),
				"remote_ip", c.RealIP(),
			)

			return err
		}
	}
}
