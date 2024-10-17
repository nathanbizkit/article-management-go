package model

import (
	"errors"
	"fmt"
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/nathanbizkit/article-management/message"
)

const (
	articleShortMaxLen = 100
	articleLongMaxLen  = 255
	tagMaxLen          = 50
)

// Tag model
type Tag struct {
	Name      string
	CreatedAt time.Time
	UpdatedAt *time.Time
}

// Article model
type Article struct {
	ID            uint
	Title         string
	Description   string
	Body          string
	Tags          []Tag
	UserID        uint
	Author        User
	FavoriteCount int64
	Comments      []Comment
	CreatedAt     time.Time
	UpdatedAt     *time.Time
}

// Validate validates fields of article model
func (a Article) Validate() error {
	return validation.ValidateStruct(&a,
		validation.Field(
			&a.Title,
			validation.Required,
			validation.Length(0, articleShortMaxLen),
		),
		validation.Field(
			&a.Description,
			validation.Length(0, articleShortMaxLen),
		),
		validation.Field(
			&a.Body,
			validation.Required,
		),
		validation.Field(
			&a.UserID,
			validation.Required,
		),
		validation.Field(
			&a.Tags,
			validation.Required,
			validation.Each(validation.By(validateTag)),
		),
	)
}

func validateTag(value interface{}) error {
	t, _ := value.(Tag)
	if t.Name == "" {
		return errors.New("must not be empty")
	}
	if len(t.Name) > tagMaxLen {
		return fmt.Errorf("the length must be no more than %d", tagMaxLen)
	}
	return nil
}

// Overwrite overwrites each field if it's not zero-value
func (a *Article) Overwrite(title, description, body string) {
	if title != "" {
		a.Title = title
	}

	if description != "" {
		a.Description = description
	}

	if body != "" {
		a.Body = body
	}
}

// ResponseArticle generates response message from article
func (a *Article) ResponseArticle(favorited bool, author message.ProfileResponse) message.ArticleResponse {
	ar := message.ArticleResponse{
		ID:             a.ID,
		Title:          a.Title,
		Description:    a.Description,
		Body:           a.Body,
		Favorited:      favorited,
		FavoritesCount: a.FavoriteCount,
		Author:         author,
		CreatedAt:      a.CreatedAt.Format(time.RFC3339Nano),
	}

	if a.UpdatedAt != nil {
		d := a.UpdatedAt.Format(time.RFC3339Nano)
		ar.UpdatedAt = &d
	}

	tags := make([]string, 0, len(a.Tags))
	for _, t := range a.Tags {
		tags = append(tags, t.Name)
	}
	ar.Tags = tags

	return ar
}
