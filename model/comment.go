package model

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/nathanbizkit/article-management-go/message"
)

// Comment model
type Comment struct {
	ID        uint
	Body      string
	UserID    uint
	Author    User
	ArticleID uint
	CreatedAt time.Time
	UpdatedAt time.Time
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

// ResponseComment generates response message for comment
func (c *Comment) ResponseComment(followingAuthor bool) message.CommentResponse {
	return message.CommentResponse{
		ID:        c.ID,
		Body:      c.Body,
		Author:    c.Author.ResponseProfile(followingAuthor),
		CreatedAt: c.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt: c.UpdatedAt.Format(time.RFC3339Nano),
	}
}
