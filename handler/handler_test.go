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

func setup(t *testing.T) (*Handler, *container.LocalTestContainer) {
	t.Helper()

	l := test.NewTestLogger(t)
	environ := test.NewTestENV(t)
	lct := test.NewLocalTestContainer(t)

	authen := auth.New(environ)
	as := store.NewArticleStore(lct.DB())
	us := store.NewUserStore(lct.DB())

	return New(&l, environ, authen, us, as), lct
}

func ctxWithToken(t *testing.T, w http.ResponseWriter, req *http.Request, id uint, timeNow time.Time) (*gin.Context, *auth.AuthToken) {
	t.Helper()

	environ := test.NewTestENV(t)
	authen := auth.New(environ)

	token, err := authen.GenerateTokenWithTime(id, timeNow)
	if err != nil {
		t.Fatal(err)
	}

	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req.Clone(context.Background())

	test.AddCookieToRequest(
		t, ctx.Request,
		"session", token.Token, "localhost",
	)
	test.AddCookieToRequest(
		t, ctx.Request,
		"refreshToken", token.RefreshToken, "localhost",
	)

	authen.SetContextUserID(ctx, id)

	return ctx, token
}

func createRandomUser(t *testing.T, db *sql.DB) *model.User {
	t.Helper()

	randStr := test.RandomString(t, 10)
	m := model.User{
		Username: fmt.Sprintf("user_%s", randStr),
		Email:    fmt.Sprintf("%s@example.com", randStr),
		Password: userPassword,
		Name:     fmt.Sprintf("USER %s", strings.ToUpper(randStr)),
		Bio:      "This is my bio.",
		Image:    "https://imgur.com/image.jpg",
	}
	m.HashPassword()

	us := store.NewUserStore(db)
	user, err := us.Create(context.Background(), &m)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		deleteUser(t, db, user.ID)
	})

	return user
}

func deleteUser(t *testing.T, db *sql.DB, id uint) {
	t.Helper()

	queryString := `DELETE FROM article_management.users WHERE id = $1`
	_, err := db.Exec(queryString, id)
	if err != nil {
		t.Fatal(err)
	}
}

func createRandomArticle(t *testing.T, db *sql.DB, userID uint, tagNames []string) *model.Article {
	t.Helper()

	randStr := test.RandomString(t, 15)
	m := model.Article{
		Title:       randStr,
		Description: randStr,
		Body:        randStr,
		UserID:      userID,
	}

	if len(tagNames) != 0 {
		tags := make([]model.Tag, 0, len(tagNames))
		for _, name := range tagNames {
			tags = append(tags, model.Tag{Name: name})
		}

		m.Tags = tags
	}

	as := store.NewArticleStore(db)
	article, err := as.Create(context.Background(), &m)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		deleteArticle(t, db, article.ID)
	})

	return article
}

func deleteArticle(t *testing.T, db *sql.DB, id uint) {
	t.Helper()

	as := store.NewArticleStore(db)
	err := as.Delete(context.Background(), &model.Article{ID: id})
	if err != nil {
		t.Fatal(err)
	}
}

func createRandomComment(t *testing.T, db *sql.DB, articleID uint, userID uint) *model.Comment {
	t.Helper()

	randStr := test.RandomString(t, 20)
	m := model.Comment{
		Body:      randStr,
		ArticleID: articleID,
		UserID:    userID,
	}

	as := store.NewArticleStore(db)
	comment, err := as.CreateComment(context.Background(), &m)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		deleteComment(t, db, comment.ID)
	})

	return comment
}

func deleteComment(t *testing.T, db *sql.DB, id uint) {
	t.Helper()

	as := store.NewArticleStore(db)
	err := as.DeleteComment(context.Background(), &model.Comment{ID: id})
	if err != nil {
		t.Fatal(err)
	}
}
