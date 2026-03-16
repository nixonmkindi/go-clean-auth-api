package controllers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/nixonmkindi/go-clean-auth-api/internal/http/middleware"
	"github.com/nixonmkindi/go-clean-auth-api/internal/http/requests"
	"github.com/nixonmkindi/go-clean-auth-api/internal/http/responses"
	"github.com/nixonmkindi/go-clean-auth-api/internal/usecase"
)

type APIClientController struct {
	apiClientUsecase *usecase.APIClientUsecase
}

func NewAPIClientController(apiClientUsecase *usecase.APIClientUsecase) *APIClientController {
	return &APIClientController{apiClientUsecase: apiClientUsecase}
}

func (ctl *APIClientController) Register(c echo.Context) error {
	var req requests.RegisterAPIClientRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	ownerID, err := middleware.GetUserID(c)
	if err != nil {
		return responses.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized", nil)
	}

	client, secret, err := ctl.apiClientUsecase.Register(c.Request().Context(), usecase.RegisterAPIClientInput{
		Name:    req.Name,
		OwnerID: ownerID,
	})
	if err != nil {
		return err
	}

	return responses.Success(c, http.StatusCreated, map[string]interface{}{
		"id":         client.ID,
		"name":       client.Name,
		"api_key":    client.Key,
		"api_secret": secret,
		"created_at": client.CreatedAt,
	})
}
