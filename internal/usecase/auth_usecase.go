package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nixonmkindi/go-clean-auth-api/internal/domain"
	"github.com/nixonmkindi/go-clean-auth-api/internal/pkg/apperror"
	"github.com/nixonmkindi/go-clean-auth-api/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthUsecase struct {
	users      repository.UserRepository
	apiClients repository.APIClientRepository
	jwtSecret  string
	jwtTTL     time.Duration
}

type RegisterUserInput struct {
	Name     string
	Email    string
	Password string
}

type LoginInput struct {
	Email    string
	Password string
}

func NewAuthUsecase(users repository.UserRepository, apiClients repository.APIClientRepository, jwtSecret string, jwtTTLMinutes int) *AuthUsecase {
	return &AuthUsecase{
		users:      users,
		apiClients: apiClients,
		jwtSecret:  jwtSecret,
		jwtTTL:     time.Duration(jwtTTLMinutes) * time.Minute,
	}
}

func (u *AuthUsecase) RegisterUser(ctx context.Context, input RegisterUserInput) (*domain.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperror.New(500, "PASSWORD_HASH_ERROR", "failed to hash password", nil, err)
	}

	user := &domain.User{
		Name:         input.Name,
		Email:        strings.ToLower(strings.TrimSpace(input.Email)),
		PasswordHash: string(hash),
	}

	if err := u.users.Create(ctx, user); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, apperror.New(409, "USER_EXISTS", "user with this email already exists", nil, err)
		}
		return nil, apperror.New(500, "USER_CREATE_ERROR", "failed to create user", nil, err)
	}

	user.PasswordHash = ""
	return user, nil
}

func (u *AuthUsecase) Login(ctx context.Context, input LoginInput) (string, time.Time, *domain.User, error) {
	user, err := u.users.GetByEmail(ctx, strings.ToLower(strings.TrimSpace(input.Email)))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", time.Time{}, nil, apperror.New(401, "INVALID_CREDENTIALS", "invalid email or password", nil, err)
		}
		return "", time.Time{}, nil, apperror.New(500, "AUTH_LOOKUP_ERROR", "failed to authenticate", nil, err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return "", time.Time{}, nil, apperror.New(401, "INVALID_CREDENTIALS", "invalid email or password", nil, err)
	}

	token, expiresAt, err := u.generateJWT(user)
	if err != nil {
		return "", time.Time{}, nil, apperror.New(500, "TOKEN_GENERATION_ERROR", "failed to generate token", nil, err)
	}

	user.PasswordHash = ""
	return token, expiresAt, user, nil
}

func (u *AuthUsecase) ValidateAPIClient(ctx context.Context, key, secret string) (*domain.APIClient, error) {
	client, err := u.apiClients.GetByKey(ctx, key)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.New(401, "INVALID_API_CREDENTIALS", "invalid api key or secret", nil, err)
		}
		return nil, apperror.New(500, "API_CLIENT_LOOKUP_ERROR", "failed to validate api client", nil, err)
	}

	if !client.IsActive {
		return nil, apperror.New(403, "API_CLIENT_INACTIVE", "api client is inactive", nil, nil)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(client.SecretHash), []byte(secret)); err != nil {
		return nil, apperror.New(401, "INVALID_API_CREDENTIALS", "invalid api key or secret", nil, err)
	}

	return client, nil
}

func (u *AuthUsecase) generateJWT(user *domain.User) (string, time.Time, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(u.jwtTTL)

	claims := jwt.MapClaims{
		"sub":   fmt.Sprintf("%d", user.ID),
		"email": user.Email,
		"name":  user.Name,
		"iat":   now.Unix(),
		"exp":   expiresAt.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(u.jwtSecret))
	if err != nil {
		return "", time.Time{}, err
	}

	return signedToken, expiresAt, nil
}
