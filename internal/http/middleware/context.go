package middleware

import (
	"fmt"
	"strconv"

	"github.com/labstack/echo/v4"
)

const (
	ContextUserIDKey      = "user_id"
	ContextAPIClientIDKey = "api_client_id"
	ContextAPIOwnerIDKey  = "api_owner_id"
	headerAPIKey          = "X-API-Key"
	headerAPISecret       = "X-API-Secret"
)

func GetUserID(c echo.Context) (int64, error) {
	value := c.Get(ContextUserIDKey)
	if value == nil {
		return 0, fmt.Errorf("user id not found in context")
	}
	id, ok := value.(int64)
	if !ok {
		return 0, fmt.Errorf("invalid user id type")
	}
	return id, nil
}

func GetAPIOwnerID(c echo.Context) (int64, error) {
	value := c.Get(ContextAPIOwnerIDKey)
	if value == nil {
		return 0, fmt.Errorf("api owner id not found in context")
	}
	id, ok := value.(int64)
	if !ok {
		return 0, fmt.Errorf("invalid api owner id type")
	}
	return id, nil
}

func ParseIDParam(c echo.Context, name string) (int64, error) {
	return ParseInt64(c.Param(name), name)
}

func ParseInt64(raw, field string) (int64, error) {
	if raw == "" {
		return 0, fmt.Errorf("%s is required", field)
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s", field)
	}
	return id, nil
}
