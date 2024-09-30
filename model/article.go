package model

import (
	"errors"
	"fmt"
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
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
	ID             uint
	Title          string
	Description    string
	Body           string
	Tags           []Tag
	UserID         uint
	Author         User
	FavoriteCount  int64
	FavoritedUsers []User
	Comments       []Comment
	CreatedAt      time.Time
	UpdatedAt      *time.Time
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
	if len(t.Name) == 0 {
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
