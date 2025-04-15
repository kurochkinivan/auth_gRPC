package entity

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID
	Email     string
	PassHash  []byte
	CreatedAt time.Time
	UpdatedAt time.Time
}
