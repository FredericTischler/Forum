package models

import (
	"time"
)

type Comment struct {
	ID                  int
	UserID              User
	PostID              int
	Content             string
	LikeDislikeComment  []LikeDislikeComment
	CreatedAt           time.Time
	LikeCountComment    int
	DislikeCountComment int
	UserAction          string
}

type CommentActivity struct {
	ID                  int
	UserID              User
	PostID              Post
	Content             string
	LikeDislikeComment  []LikeDislikeComment
	CreatedAt           time.Time
	LikeCountComment    int
	DislikeCountComment int
	UserAction          string
}
