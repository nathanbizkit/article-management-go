package model

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/nathanbizkit/article-management/message"
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

// ResponseComment generates response message for comment
func (c *Comment) ResponseComment(followingAuthor bool) message.CommentResponse {
	cr := message.CommentResponse{
		ID:   c.ID,
		Body: c.Body,
		Author: message.ProfileResponse{
			Username:  c.Author.Username,
			Bio:       c.Author.Bio,
			Image:     c.Author.Image,
			Following: followingAuthor,
		},
		CreatedAt: c.CreatedAt.Format(time.RFC3339Nano),
	}

	if c.UpdatedAt != nil {
		d := c.UpdatedAt.Format(time.RFC3339Nano)
		cr.UpdatedAt = &d
	}

	return cr
}
