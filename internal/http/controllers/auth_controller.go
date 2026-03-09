package controllers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/yourusername/go-clean-auth-api/internal/http/requests"
	"github.com/yourusername/go-clean-auth-api/internal/http/responses"
	"github.com/yourusername/go-clean-auth-api/internal/usecase"
)

type AuthController struct {
	authUsecase *usecase.AuthUsecase
}

func NewAuthController(authUsecase *usecase.AuthUsecase) *AuthController {
	return &AuthController{authUsecase: authUsecase}
}

func (ctl *AuthController) Register(c echo.Context) error {
	var req requests.RegisterUserRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	user, err := ctl.authUsecase.RegisterUser(c.Request().Context(), usecase.RegisterUserInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return err
	}

	return responses.Success(c, http.StatusCreated, user)
}

func (ctl *AuthController) Login(c echo.Context) error {
	var req requests.LoginRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	token, expiresAt, user, err := ctl.authUsecase.Login(c.Request().Context(), usecase.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return err
	}

	return responses.Success(c, http.StatusOK, map[string]interface{}{
		"token":      token,
		"token_type": "Bearer",
		"expires_at": expiresAt,
		"user":       user,
	})
}
