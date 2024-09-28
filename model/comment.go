package model

import "time"

type Comment struct {
	ID        uint
	Body      string
	UserID    uint
	Author    User
	ArticleID uint
	Article   Article
	CreatedAt time.Time
	UpdatedAt *time.Time
}
