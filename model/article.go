package model

import "time"

// Article model
type Article struct {
	ID             uint
	Title          string
	Description    string
	Body           string
	UserID         uint
	Author         User
	FavoriteCount  int64
	FavoritedUsers []User
	Comments       []Comment
	CreatedAt      time.Time
	UpdatedAt      *time.Time
}
