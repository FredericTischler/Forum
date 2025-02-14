package models

import (
	"time"
)

type Post struct {
	ID           int
	UserID       User
	Title        string
	Content      string
	Image        *string
	Category     []Category
	Comments     []Comment
	LikeCount    int
	DislikeCount int
	UserAction   string
	CreatedAt    time.Time
}
