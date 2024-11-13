package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
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

func ctxWithToken(t *testing.T, w http.ResponseWriter, id uint) (*gin.Context, *auth.AuthToken) {
	t.Helper()

	e := test.NewTestENV(t)
	a := auth.New(e)

	token, err := a.GenerateToken(id)
	if err != nil {
		t.Fatal(err)
	}

	c, _ := gin.CreateTestContext(w)
	c.Request = &http.Request{
		Header: make(http.Header),
	}

	a.SetContextUserID(c, id)

	addCookieToRequest(t, c.Request, "session",
		token.Token, "localhost")
	addCookieToRequest(t, c.Request, "refreshToken",
		token.RefreshToken, "localhost")

	return c, token
}

func addCookieToRequest(t *testing.T, req *http.Request, name, value, domain string) {
	t.Helper()

	cookie := &http.Cookie{
		Name:     name,
		Value:    url.QueryEscape(value),
		MaxAge:   int((20 * (25 * time.Hour)).Seconds()),
		Path:     "/api",
		Domain:   domain,
		SameSite: http.SameSiteStrictMode,
		Secure:   true,
		HttpOnly: true,
	}

	if v := cookie.String(); v != "" {
		req.Header.Add("Cookie", v)
	}
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
