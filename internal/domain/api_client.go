package domain

import "time"

type APIClient struct {
	ID         int64      `json:"id"`
	Name       string     `json:"name"`
	Key        string     `json:"key"`
	SecretHash string     `json:"-"`
	OwnerID    int64      `json:"owner_id"`
	IsActive   bool       `json:"is_active"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`
}
