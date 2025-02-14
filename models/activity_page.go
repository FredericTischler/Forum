package models

import "time"

type ActivityPage struct {
	Id           int
	UserID       User
	ActivityType string
	PostID       *Post
	CommentID    *CommentActivity
	CreatedAt    time.Time
}
