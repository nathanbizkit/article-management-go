package model

import (
	"fmt"
	"testing"
	"time"

	"github.com/nathanbizkit/article-management/message"
	"github.com/stretchr/testify/assert"
)

func TestCommentModel_Validate(t *testing.T) {
	tests := []struct {
		title    string
		c        *Comment
		hasError bool
	}{
		{
			"validate comment: success",
			&Comment{
				Body:      "A text body.",
				ArticleID: 1,
				UserID:    1,
			},
			false,
		},
		{
			"validate comment: no body",
			&Comment{
				Body:      "",
				ArticleID: 1,
				UserID:    1,
			},
			true,
		},
		{
			"validate comment: no article id",
			&Comment{
				Body:      "A text body.",
				ArticleID: 0,
				UserID:    1,
			},
			true,
		},
		{
			"validate comment: no user id",
			&Comment{
				Body:      "A text body.",
				ArticleID: 1,
				UserID:    0,
			},
			true,
		},
	}

	for _, tt := range tests {
		err := tt.c.Validate()
		if tt.hasError {
			assert.Error(t, err, fmt.Sprintf("%s: expect an error", tt.title))
		} else {
			assert.NoError(t, err, fmt.Sprintf("%s: expect no error", tt.title))
		}
	}
}

func TestCommentModel_ResponseComment(t *testing.T) {
	createdAt := time.Now()
	updatedAt := time.Now().Add(10 * time.Hour)
	updatedAtString := updatedAt.Format(time.RFC3339Nano)

	c := Comment{
		ID:     1,
		Body:   "This is my comment.",
		UserID: 1,
		Author: User{
			ID:        1,
			Username:  "foo_user",
			Email:     "foo@example.com",
			Password:  "encrypted_password",
			Name:      "FooUser",
			Bio:       "This is my bio.",
			Image:     "https://imgur.com/image.jpeg",
			CreatedAt: time.Now(),
			UpdatedAt: nil,
		},
		ArticleID: 1,
		CreatedAt: createdAt,
		UpdatedAt: &updatedAt,
	}

	expected := message.CommentResponse{
		ID:   1,
		Body: "This is my comment.",
		Author: message.ProfileResponse{
			Username:  "foo_user",
			Name:      "FooUser",
			Bio:       "This is my bio.",
			Image:     "https://imgur.com/image.jpeg",
			Following: false,
		},
		CreatedAt: createdAt.Format(time.RFC3339Nano),
		UpdatedAt: &updatedAtString,
	}

	cr := c.ResponseComment(false)
	assert.Equal(t, expected, cr)
}
