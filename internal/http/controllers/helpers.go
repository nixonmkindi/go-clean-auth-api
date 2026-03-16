package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/nixonmkindi/go-clean-auth-api/internal/http/responses"
)

func bindAndValidate(c echo.Context, req interface{}) error {
	if err := c.Bind(req); err != nil {
		return responses.Fail(c, http.StatusBadRequest, "INVALID_JSON", "invalid request body", nil)
	}
	if err := c.Validate(req); err != nil {
		details := responses.BuildValidationDetails(err)
		return responses.ValidationError(c, details)
	}
	return nil
}

func getQueryInt(c echo.Context, key string, defaultValue int) int {
	raw := c.QueryParam(key)
	if raw == "" {
		return defaultValue
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return defaultValue
	}
	return v
}

func parseOptionalRFC3339(value *string) (*time.Time, error) {
	if value == nil || *value == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, *value)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
