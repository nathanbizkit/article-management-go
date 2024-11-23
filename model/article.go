package model

import (
	"errors"
	"fmt"
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/nathanbizkit/article-management/message"
)

const (
	articleShortMinLen = 5
	articleShortMaxLen = 100
	tagMinLen          = 3
	tagMaxLen          = 50
)

// Tag model
type Tag struct {
	ID        uint
	Name      string
	CreatedAt time.Time
	UpdatedAt *time.Time
}

// Article model
type Article struct {
	ID             uint
	Title          string
	Description    string
	Body           string
	Tags           []Tag
	UserID         uint
	Author         User
	FavoritesCount int64
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// Validate validates fields of article model
func (a Article) Validate() error {
	return validation.ValidateStruct(&a,
		validation.Field(
			&a.Title,
			validation.Required,
			validation.Length(articleShortMinLen, articleShortMaxLen),
		),
		validation.Field(
			&a.Description,
			validation.Length(articleShortMinLen, articleShortMaxLen),
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
		return errors.New("tag name cannot be blank")
	}

	if len(t.Name) < tagMinLen || len(t.Name) > tagMaxLen {
		return fmt.Errorf("tag name length must be between %d and %d", tagMinLen, tagMaxLen)
	}

	return nil
}

// Overwrite overwrites each field if it's not zero-value
func (a *Article) Overwrite(title, description, body string) {
	if title != "" {
		a.Title = title
	}

	if body != "" {
		a.Body = body
	}

	a.Description = description
}

// ResponseArticle generates response message from article
func (a *Article) ResponseArticle(favorited, followingAuthor bool) message.ArticleResponse {
	resp := message.ArticleResponse{
		ID:             a.ID,
		Title:          a.Title,
		Description:    a.Description,
		Body:           a.Body,
		Favorited:      favorited,
		FavoritesCount: a.FavoritesCount,
		Author:         a.Author.ResponseProfile(followingAuthor),
		CreatedAt:      a.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:      a.UpdatedAt.Format(time.RFC3339Nano),
	}

	tags := make([]string, 0, len(a.Tags))
	for _, t := range a.Tags {
		tags = append(tags, t.Name)
	}

	resp.Tags = tags
	return resp
}
