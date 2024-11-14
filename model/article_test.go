package model

import (
	"strings"
	"testing"
	"time"

	"github.com/nathanbizkit/article-management/message"
	"github.com/stretchr/testify/assert"
)

func TestUnit_ArticleModel(t *testing.T) {
	if !testing.Short() {
		t.Skip("skipping unit tests.")
	}

	t.Run("Validate", func(t *testing.T) {
		shortMaxLenString := strings.Repeat("a", 101)
		tagMaxLenString := strings.Repeat("a", 51)

		tests := []struct {
			title    string
			a        *Article
			hasError bool
		}{
			{
				"validate article: success",
				&Article{
					Title:       "Article 1",
					Description: "This is a description.",
					Body:        "This is a text body.",
					UserID:      1,
					Tags:        []Tag{{Name: "tag-1"}, {Name: "tag-2"}},
				},
				false,
			},
			{
				"validate article: no title",
				&Article{
					Title:       "",
					Description: "This is a description.",
					Body:        "This is a text body.",
					UserID:      1,
					Tags:        []Tag{{Name: "tag-1"}, {Name: "tag-2"}},
				},
				true,
			},
			{
				"validate article: title is too short",
				&Article{
					Title:       "Art",
					Description: "This is a description.",
					Body:        "This is a text body.",
					UserID:      1,
					Tags:        []Tag{{Name: "tag-1"}, {Name: "tag-2"}},
				},
				true,
			},
			{
				"validate article: title is too long",
				&Article{
					Title:       shortMaxLenString,
					Description: "This is a description.",
					Body:        "This is a text body.",
					UserID:      1,
					Tags:        []Tag{{Name: "tag-1"}, {Name: "tag-2"}},
				},
				true,
			},
			{
				"validate article: description is too short",
				&Article{
					Title:       "Article 1",
					Description: "This",
					Body:        "This is a text body.",
					UserID:      1,
					Tags:        []Tag{{Name: "tag-1"}, {Name: "tag-2"}},
				},
				true,
			},
			{
				"validate article: description is too long",
				&Article{
					Title:       "Article 1",
					Description: shortMaxLenString,
					Body:        "This is a text body.",
					UserID:      1,
					Tags:        []Tag{{Name: "tag-1"}, {Name: "tag-2"}},
				},
				true,
			},
			{
				"validate article: no body",
				&Article{
					Title:       "Article 1",
					Description: "This is a description.",
					Body:        "",
					UserID:      1,
					Tags:        []Tag{{Name: "tag-1"}, {Name: "tag-2"}},
				},
				true,
			},
			{
				"validate article: no user id",
				&Article{
					Title:       "Article 1",
					Description: "This is a description.",
					Body:        "This is a text body.",
					UserID:      0,
					Tags:        []Tag{{Name: "tag-1"}, {Name: "tag-2"}},
				},
				true,
			},
			{
				"validate article: no tags",
				&Article{
					Title:       "Article 1",
					Description: "This is a description.",
					Body:        "This is a text body.",
					UserID:      1,
					Tags:        []Tag{},
				},
				true,
			},
			{
				"validate article: empty tag name",
				&Article{
					Title:       "Article 1",
					Description: "This is a description.",
					Body:        "This is a text body.",
					UserID:      1,
					Tags:        []Tag{{Name: "tag-1"}, {Name: ""}},
				},
				true,
			},
			{
				"validate article: tag name is too short",
				&Article{
					Title:       "Article 1",
					Description: "This is a description.",
					Body:        "This is a text body.",
					UserID:      1,
					Tags:        []Tag{{Name: "a"}, {Name: "b"}},
				},
				true,
			},
			{
				"validate article: tag name is too long",
				&Article{
					Title:       "Article 1",
					Description: "This is a description.",
					Body:        "This is a text body.",
					UserID:      1,
					Tags:        []Tag{{Name: "tag-1"}, {Name: tagMaxLenString}},
				},
				true,
			},
		}

		for _, tt := range tests {
			err := tt.a.Validate()

			if tt.hasError {
				assert.Error(t, err, tt.title)
			} else {
				assert.NoError(t, err, tt.title)
			}
		}
	})

	t.Run("Overwrite", func(t *testing.T) {
		now := time.Now()

		tests := []struct {
			title    string
			a        *Article
			in       *Article
			expected *Article
		}{
			{
				"overwrite article: success",
				&Article{
					ID:          1,
					Title:       "Article 1",
					Description: "This is a description.",
					Body:        "This is a text body.",
					UserID:      1,
					Tags:        []Tag{{Name: "tag-1"}, {Name: "tag-2"}},
					CreatedAt:   now,
				},
				&Article{
					Title:       "New Article 1",
					Description: "This is a new description.",
					Body:        "This is a new text body.",
				},
				&Article{
					ID:          1,
					Title:       "New Article 1",
					Description: "This is a new description.",
					Body:        "This is a new text body.",
					UserID:      1,
					Tags:        []Tag{{Name: "tag-1"}, {Name: "tag-2"}},
					CreatedAt:   now,
				},
			},
			{
				"overwrite article: empty description and no changes for other fields",
				&Article{
					ID:          1,
					Title:       "Article 1",
					Description: "This is a description.",
					Body:        "This is a text body.",
					UserID:      1,
					Tags:        []Tag{{Name: "tag-1"}, {Name: "tag-2"}},
					CreatedAt:   now,
				},
				&Article{
					Description: "",
				},
				&Article{
					ID:          1,
					Title:       "Article 1",
					Description: "",
					Body:        "This is a text body.",
					UserID:      1,
					Tags:        []Tag{{Name: "tag-1"}, {Name: "tag-2"}},
					CreatedAt:   now,
				},
			},
		}

		for _, tt := range tests {
			tt.a.Overwrite(tt.in.Title, tt.in.Description, tt.in.Body)
			assert.Equal(t, tt.expected, tt.a, tt.title)
		}
	})

	t.Run("ResponseArticle", func(t *testing.T) {
		createdAt := time.Now()
		updatedAt := time.Now().Add(10 * time.Hour)
		updatedAtString := updatedAt.Format(time.RFC3339Nano)

		favorited := false
		following := false
		expected := message.ArticleResponse{
			ID:          1,
			Title:       "Article 1",
			Description: "This is a description.",
			Body:        "This is a text body.",
			Author: message.ProfileResponse{
				Username:  "foo_user",
				Name:      "FooUser",
				Bio:       "This is my bio.",
				Image:     "https://imgur.com/image.jpeg",
				Following: following,
			},
			Favorited:      favorited,
			FavoritesCount: 10,
			CreatedAt:      createdAt.Format(time.RFC3339Nano),
			UpdatedAt:      &updatedAtString,
			Tags:           []string{"tag-1", "tag-2"},
		}

		a := Article{
			ID:          1,
			Title:       "Article 1",
			Description: "This is a description.",
			Body:        "This is a text body.",
			UserID:      1,
			Author: User{
				ID:        1,
				Username:  "foo_user",
				Email:     "foo@example.com",
				Password:  "encrypted_password",
				Name:      "FooUser",
				Bio:       "This is my bio.",
				Image:     "https://imgur.com/image.jpeg",
				CreatedAt: createdAt,
				UpdatedAt: nil,
			},
			FavoritesCount: 10,
			Tags:           []Tag{{Name: "tag-1"}, {Name: "tag-2"}},
			CreatedAt:      createdAt,
			UpdatedAt:      &updatedAt,
		}

		actual := a.ResponseArticle(favorited, following)
		assert.Equal(t, expected, actual)
	})
}
