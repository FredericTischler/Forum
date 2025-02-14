package models

import (
	"time"

	"github.com/google/uuid"
)

type Report struct {
	ID        int
	UserID    uuid.UUID
	PostID    *int // Pointeur pour permettre null (si c'est un commentaire signalé)
	CommentID *int // Pointeur pour permettre null (si c'est un post signalé)
	Reason    string
	CreatedAt time.Time
}
