package model

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
)

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

// Validate validates fields of comment model
func (c Comment) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(
			&c.Body,
			validation.Required,
		),
		validation.Field(
			&c.UserID,
			validation.Required,
		),
		validation.Field(
			&c.ArticleID,
			validation.Required,
		),
	)
}
