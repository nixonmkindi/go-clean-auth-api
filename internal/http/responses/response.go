package responses

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/nixonmkindi/go-clean-auth-api/internal/pkg/apperror"
)

type Envelope struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorBody  `json:"error,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
}

type ErrorBody struct {
	Code    string      `json:"code"`
	Details interface{} `json:"details,omitempty"`
}

type ValidationFieldError struct {
	Field string `json:"field"`
	Tag   string `json:"tag"`
}

func Success(c echo.Context, status int, data interface{}) error {
	return c.JSON(status, Envelope{Success: true, Data: data})
}

func Message(c echo.Context, status int, message string) error {
	return c.JSON(status, Envelope{Success: true, Message: message})
}

func Paginated(c echo.Context, status int, data interface{}, meta interface{}) error {
	return c.JSON(status, Envelope{Success: true, Data: data, Meta: meta})
}

func ValidationError(c echo.Context, details interface{}) error {
	return c.JSON(http.StatusBadRequest, Envelope{
		Success: false,
		Message: "validation failed",
		Error: &ErrorBody{
			Code:    "VALIDATION_ERROR",
			Details: details,
		},
	})
}

func Fail(c echo.Context, status int, code, message string, details interface{}) error {
	return c.JSON(status, Envelope{
		Success: false,
		Message: message,
		Error: &ErrorBody{
			Code:    code,
			Details: details,
		},
	})
}

func BuildValidationDetails(err error) []ValidationFieldError {
	var validationErrs validator.ValidationErrors
	if !errors.As(err, &validationErrs) {
		return []ValidationFieldError{{Field: "request", Tag: err.Error()}}
	}

	details := make([]ValidationFieldError, 0, len(validationErrs))
	for _, e := range validationErrs {
		details = append(details, ValidationFieldError{
			Field: e.Field(),
			Tag:   e.Tag(),
		})
	}

	return details
}

func HTTPErrorHandler(log *slog.Logger) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		var appErr *apperror.Error
		if errors.As(err, &appErr) {
			_ = Fail(c, appErr.HTTPStatus, appErr.Code, appErr.Message, appErr.Details)
			return
		}

		var httpErr *echo.HTTPError
		if errors.As(err, &httpErr) {
			code := "HTTP_ERROR"
			message := "request failed"
			if msg, ok := httpErr.Message.(string); ok {
				message = msg
			}
			_ = Fail(c, httpErr.Code, code, message, nil)
			return
		}

		log.Error("unhandled_error", "error", err.Error(), "path", c.Path(), "method", c.Request().Method)
		_ = Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error", nil)
	}
}
