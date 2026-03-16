package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nixonmkindi/go-clean-auth-api/internal/domain"
)

type APIClientRepository struct {
	pool *pgxpool.Pool
}

func NewAPIClientRepository(pool *pgxpool.Pool) *APIClientRepository {
	return &APIClientRepository{pool: pool}
}

func (r *APIClientRepository) Create(ctx context.Context, client *domain.APIClient) error {
	query := `
		INSERT INTO api_clients (name, api_key, secret_hash, owner_id, is_active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return r.pool.QueryRow(ctx, query, client.Name, client.Key, client.SecretHash, client.OwnerID, client.IsActive).
		Scan(&client.ID, &client.CreatedAt, &client.UpdatedAt)
}

func (r *APIClientRepository) GetByKey(ctx context.Context, key string) (*domain.APIClient, error) {
	query := `
		SELECT id, name, api_key, secret_hash, owner_id, is_active, created_at, updated_at, deleted_at
		FROM api_clients
		WHERE api_key = $1 AND deleted_at IS NULL`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var client domain.APIClient
	err := r.pool.QueryRow(ctx, query, key).Scan(
		&client.ID,
		&client.Name,
		&client.Key,
		&client.SecretHash,
		&client.OwnerID,
		&client.IsActive,
		&client.CreatedAt,
		&client.UpdatedAt,
		&client.DeletedAt,
	)
	if err != nil {
		return nil, err
	}

	return &client, nil
}
