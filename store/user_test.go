package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/nathanbizkit/article-management/model"
	"github.com/nathanbizkit/article-management/test"
	"github.com/stretchr/testify/assert"
)

func TestIntegration_UserStore(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration tests.")
	}

	lct := test.NewLocalTestContainer(t)
	us := NewUserStore(lct.DB())
	u := createUser(t, lct.DB())

	t.Run("GetByID", func(t *testing.T) {
		tests := []struct {
			title    string
			in       uint
			expected *model.User
			hasError bool
		}{
			{
				"get by id user: success",
				u.ID,
				u,
				false,
			},
			{
				"get by id user: not found",
				999,
				nil,
				true,
			},
		}

		for _, tt := range tests {
			actual, err := us.GetByID(context.Background(), tt.in)

			if tt.hasError {
				assert.Error(t, err, tt.title)
			} else {
				assert.NoError(t, err, tt.title)
			}

			assert.Equal(t, tt.expected, actual, tt.title)
		}
	})
}

func createUser(t *testing.T, db *sql.DB) *model.User {
	t.Helper()

	s := test.RandomString(t, 10)
	m := model.User{
		Username: fmt.Sprintf("user_%s", s),
		Email:    fmt.Sprintf("%s@example.com", s),
		Password: "P@55w0rD!",
		Name:     fmt.Sprintf("USER %s", strings.ToUpper(s)),
		Bio:      "This is my bio.",
		Image:    "https://imgur.com/image.jpeg",
	}
	m.HashPassword()

	var u model.User
	queryString := `INSERT INTO article_management.users (username, email, password, name, bio, image) 
			VALUES ($1, $2, $3, $4, $5, $6) 
			RETURNING id, username, email, password, name, bio, image, created_at, updated_at`
	err := db.QueryRow(queryString, m.Username, m.Email, m.Password, m.Name, m.Bio, m.Image).
		Scan(
			&u.ID,
			&u.Username,
			&u.Email,
			&u.Password,
			&u.Name,
			&u.Bio,
			&u.Image,
			&u.CreatedAt,
			&u.UpdatedAt,
		)
	if err != nil {
		t.Fatalf("failed to create test user: %s", err)
	}

	t.Cleanup(func() {
		deleteUser(t, db, u.ID)
	})

	return &u
}

func deleteUser(t *testing.T, db *sql.DB, id uint) {
	t.Helper()

	queryString := `DELETE FROM article_management.users WHERE id = $1`
	_, err := db.Exec(queryString, id)
	if err != nil {
		t.Fatalf("failed to delete test user: %s", err)
	}
}
