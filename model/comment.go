package model

import "time"

// Comment model
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
