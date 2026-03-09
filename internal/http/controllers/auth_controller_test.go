package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/go-clean-auth-api/internal/domain"
	customvalidator "github.com/yourusername/go-clean-auth-api/internal/pkg/validator"
	"github.com/yourusername/go-clean-auth-api/internal/usecase"
	"golang.org/x/crypto/bcrypt"
)

type authUserRepoMock struct {
	createFn     func(ctx context.Context, user *domain.User) error
	getByEmailFn func(ctx context.Context, email string) (*domain.User, error)
	getByIDFn    func(ctx context.Context, id int64) (*domain.User, error)
}

func (m *authUserRepoMock) Create(ctx context.Context, user *domain.User) error {
	if m.createFn != nil {
		return m.createFn(ctx, user)
	}
	return nil
}

func (m *authUserRepoMock) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	if m.getByEmailFn != nil {
		return m.getByEmailFn(ctx, email)
	}
	return nil, pgx.ErrNoRows
}

func (m *authUserRepoMock) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, pgx.ErrNoRows
}

type authAPIClientRepoMock struct{}

func (m *authAPIClientRepoMock) Create(ctx context.Context, client *domain.APIClient) error {
	return nil
}

func (m *authAPIClientRepoMock) GetByKey(ctx context.Context, key string) (*domain.APIClient, error) {
	return nil, pgx.ErrNoRows
}

func TestAuthControllerLoginSuccess(t *testing.T) {
	e := echo.New()
	e.Validator = customvalidator.New()

	hash, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	users := &authUserRepoMock{
		getByEmailFn: func(ctx context.Context, email string) (*domain.User, error) {
			return &domain.User{ID: 1, Name: "Alice", Email: email, PasswordHash: string(hash)}, nil
		},
	}
	authUsecase := usecase.NewAuthUsecase(users, &authAPIClientRepoMock{}, "jwt-secret", 60)
	controller := NewAuthController(authUsecase)

	e.POST("/login", controller.Login)

	payload := map[string]string{"email": "alice@example.com", "password": "secret123"}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "token")
}

func TestAuthControllerRegisterValidation(t *testing.T) {
	e := echo.New()
	e.Validator = customvalidator.New()

	authUsecase := usecase.NewAuthUsecase(&authUserRepoMock{}, &authAPIClientRepoMock{}, "jwt-secret", 60)
	controller := NewAuthController(authUsecase)

	e.POST("/register", controller.Register)

	payload := map[string]string{"email": "invalid"}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "VALIDATION_ERROR")
}
