package usecase

import (
	"context"
	"crypto/rand"
	"encoding/base64"

	"github.com/nixonmkindi/go-clean-auth-api/internal/domain"
	"github.com/nixonmkindi/go-clean-auth-api/internal/pkg/apperror"
	"github.com/nixonmkindi/go-clean-auth-api/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type APIClientUsecase struct {
	repo repository.APIClientRepository
}

type RegisterAPIClientInput struct {
	Name    string
	OwnerID int64
}

func NewAPIClientUsecase(repo repository.APIClientRepository) *APIClientUsecase {
	return &APIClientUsecase{repo: repo}
}

func (u *APIClientUsecase) Register(ctx context.Context, input RegisterAPIClientInput) (*domain.APIClient, string, error) {
	key, err := randomToken(24)
	if err != nil {
		return nil, "", apperror.New(500, "API_KEY_GENERATION_ERROR", "failed to generate api key", nil, err)
	}
	secret, err := randomToken(32)
	if err != nil {
		return nil, "", apperror.New(500, "API_SECRET_GENERATION_ERROR", "failed to generate api secret", nil, err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", apperror.New(500, "API_SECRET_HASH_ERROR", "failed to hash api secret", nil, err)
	}

	client := &domain.APIClient{
		Name:       input.Name,
		Key:        key,
		SecretHash: string(hash),
		OwnerID:    input.OwnerID,
		IsActive:   true,
	}

	if err := u.repo.Create(ctx, client); err != nil {
		return nil, "", apperror.New(500, "API_CLIENT_CREATE_ERROR", "failed to create api client", nil, err)
	}

	client.SecretHash = ""
	return client, secret, nil
}

func randomToken(byteLen int) (string, error) {
	buf := make([]byte, byteLen)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
