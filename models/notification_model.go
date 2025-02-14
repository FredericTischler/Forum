package models

type Notification struct {
	Id         int
	UserId     User
	UserId2    User
	Post_Id    *Post
	Comment_Id *Comment
	Type       string
	IsRead     bool
	CreatedAt  string
}
