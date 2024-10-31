package store

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"testing"

	"github.com/nathanbizkit/article-management/model"
	"github.com/nathanbizkit/article-management/test"
	"github.com/stretchr/testify/assert"
)

func TestIntegration_UserStore(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration tests.")
	}

	lct := test.GetLocalTestContainer(t)
	us := NewUserStore(lct.DB())
	u := createUser(t, lct.DB())

	t.Run("GetByID", func(t *testing.T) {
		tests := []struct {
			title      string
			in         uint
			expectUser *model.User
			hasError   bool
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
			u, err := us.GetByID(context.Background(), tt.in)

			if tt.hasError {
				assert.Error(t, err, fmt.Sprintf("%s: expect an error", tt.title))
				assert.Nil(t, u, fmt.Sprintf("%s: expect no user", tt.title))
			} else {
				assert.NoError(t, err, fmt.Sprintf("%s: expect no error", tt.title))
				assert.Equal(t, tt.expectUser, u, tt.title)
			}
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
		Name:     fmt.Sprintf("User %s", s),
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
		log.Fatalf("failed to create test user: %s", err)
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
		log.Fatalf("failed to delete test user: %s", err)
	}
}
