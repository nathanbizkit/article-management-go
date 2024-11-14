package handler

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nathanbizkit/article-management/auth"
	"github.com/nathanbizkit/article-management/model"
	"github.com/nathanbizkit/article-management/store"
	"github.com/nathanbizkit/article-management/test"
	"github.com/nathanbizkit/article-management/test/container"
)

const userPassword = "P@55w0rD!"

func setUp(t *testing.T) (*Handler, *container.LocalTestContainer) {
	t.Helper()

	l := test.NewTestLogger(t)
	e := test.NewTestENV(t)
	lct := test.NewLocalTestContainer(t)

	auth := auth.New(e)
	as := store.NewArticleStore(lct.DB())
	us := store.NewUserStore(lct.DB())

	return New(&l, e, auth, us, as), lct
}

func ctxWithToken(t *testing.T, w http.ResponseWriter, req *http.Request, id uint, tokenTime time.Time) (*gin.Context, *auth.AuthToken) {
	t.Helper()

	e := test.NewTestENV(t)
	a := auth.New(e)

	token, err := a.GenerateTokenWithTime(id, tokenTime)
	if err != nil {
		t.Fatal(err)
	}

	c, _ := gin.CreateTestContext(w)
	c.Request = req.Clone(context.Background())

	a.SetContextUserID(c, id)

	test.AddCookieToRequest(t, c.Request, "session",
		token.Token, "localhost")
	test.AddCookieToRequest(t, c.Request, "refreshToken",
		token.RefreshToken, "localhost")

	return c, token
}

func createUser(t *testing.T, db *sql.DB) *model.User {
	t.Helper()

	s := test.RandomString(t, 10)
	m := model.User{
		Username: fmt.Sprintf("user_%s", s),
		Email:    fmt.Sprintf("%s@example.com", s),
		Password: userPassword,
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
