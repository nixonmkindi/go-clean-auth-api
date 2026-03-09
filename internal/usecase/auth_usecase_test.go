package usecase

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/go-clean-auth-api/internal/domain"
	"github.com/yourusername/go-clean-auth-api/internal/pkg/apperror"
	"golang.org/x/crypto/bcrypt"
)

type mockUserRepo struct {
	createFn     func(ctx context.Context, user *domain.User) error
	getByEmailFn func(ctx context.Context, email string) (*domain.User, error)
	getByIDFn    func(ctx context.Context, id int64) (*domain.User, error)
}

func (m *mockUserRepo) Create(ctx context.Context, user *domain.User) error {
	if m.createFn != nil {
		return m.createFn(ctx, user)
	}
	return nil
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	if m.getByEmailFn != nil {
		return m.getByEmailFn(ctx, email)
	}
	return nil, pgx.ErrNoRows
}

func (m *mockUserRepo) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, pgx.ErrNoRows
}

type mockAPIClientRepo struct {
	createFn   func(ctx context.Context, client *domain.APIClient) error
	getByKeyFn func(ctx context.Context, key string) (*domain.APIClient, error)
}

func (m *mockAPIClientRepo) Create(ctx context.Context, client *domain.APIClient) error {
	if m.createFn != nil {
		return m.createFn(ctx, client)
	}
	return nil
}

func (m *mockAPIClientRepo) GetByKey(ctx context.Context, key string) (*domain.APIClient, error) {
	if m.getByKeyFn != nil {
		return m.getByKeyFn(ctx, key)
	}
	return nil, pgx.ErrNoRows
}

func TestAuthUsecase_RegisterUser(t *testing.T) {
	users := &mockUserRepo{
		createFn: func(ctx context.Context, user *domain.User) error {
			user.ID = 11
			return nil
		},
	}
	apiClients := &mockAPIClientRepo{}
	u := NewAuthUsecase(users, apiClients, "secret", 60)

	user, err := u.RegisterUser(context.Background(), RegisterUserInput{
		Name:     "Alice",
		Email:    "alice@example.com",
		Password: "password123",
	})
	require.NoError(t, err)
	assert.Equal(t, int64(11), user.ID)
	assert.Equal(t, "", user.PasswordHash)
}

func TestAuthUsecase_LoginInvalidPassword(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.DefaultCost)
	require.NoError(t, err)

	users := &mockUserRepo{
		getByEmailFn: func(ctx context.Context, email string) (*domain.User, error) {
			return &domain.User{ID: 1, Name: "Alice", Email: email, PasswordHash: string(hash)}, nil
		},
	}
	apiClients := &mockAPIClientRepo{}
	u := NewAuthUsecase(users, apiClients, "secret", 60)

	_, _, _, err = u.Login(context.Background(), LoginInput{
		Email:    "alice@example.com",
		Password: "wrong-password",
	})
	require.Error(t, err)

	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 401, appErr.HTTPStatus)
	assert.Equal(t, "INVALID_CREDENTIALS", appErr.Code)
}
