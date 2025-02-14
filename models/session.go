package models

import (
	"time"

	"github.com/google/uuid"
)

// Session représente une session utilisateur dans la base de données.
type Session struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
}
