package models

import (
	"time"

	"github.com/google/uuid"
)

type LikeDislikeComment struct {
	ID        int
	UserID    uuid.UUID
	PostID    int
	Like      bool
	Dislike   bool
	CreatedAt time.Time
}
