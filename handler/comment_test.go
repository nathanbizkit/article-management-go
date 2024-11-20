package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nathanbizkit/article-management/message"
	"github.com/nathanbizkit/article-management/model"
	"github.com/nathanbizkit/article-management/test"
	"github.com/stretchr/testify/assert"
)

func TestIntegration_CommentHandler(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration tests.")
	}

	gin.SetMode("test")
	h, lct := setup(t)

	t.Run("CreateComment", func(t *testing.T) {
		fooUser := createRandomUser(t, lct.DB())
		barUser := createRandomUser(t, lct.DB())

		barArticle := createRandomArticle(
			t, lct.DB(),
			barUser.ID,
			[]string{test.RandomString(t, 10), test.RandomString(t, 10)},
		)

		randStr := test.RandomString(t, 20)
		cm := model.Comment{
			Body:   randStr,
			Author: *fooUser,
		}

		following := false
		expected := cm.ResponseComment(following)

		r := message.CreateCommentRequest{
			Body: randStr,
		}

		body, err := json.Marshal(r)
		if err != nil {
			t.Fatal(err)
		}

		apiUrl := fmt.Sprintf("/api/v1/articles/%d/comments", barArticle.ID)
		req := httptest.NewRequest(http.MethodPost, apiUrl, bytes.NewReader(body))

		w := httptest.NewRecorder()
		c, _ := ctxWithToken(t, w, req, fooUser.ID, time.Now().Add(-time.Hour))
		c.AddParam("slug", strconv.Itoa(int(barArticle.ID)))

		h.CreateComment(c)

		var actual message.CommentResponse
		err = json.NewDecoder(w.Result().Body).Decode(&actual)
		if err != nil {
			t.Fatal(err)
		}
		defer w.Result().Body.Close()

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Greater(t, actual.ID, uint(0))
		assert.Equal(t, expected.Body, actual.Body)
		assert.Equal(t, expected.Author, actual.Author)
		assert.NotEmpty(t, actual.CreatedAt)
		assert.NotEmpty(t, actual.UpdatedAt)
		assert.Equal(t, actual.CreatedAt, actual.UpdatedAt)
	})

	t.Run("GetComments", func(t *testing.T) {
		fooUser := createRandomUser(t, lct.DB())
		barUser := createRandomUser(t, lct.DB())

		barArticle := createRandomArticle(
			t, lct.DB(),
			barUser.ID,
			[]string{test.RandomString(t, 10), test.RandomString(t, 10)},
		)

		cm1 := createRandomComment(t, lct.DB(), barArticle.ID, fooUser.ID)
		time.Sleep(1 * time.Second)
		cm2 := createRandomComment(t, lct.DB(), barArticle.ID, barUser.ID)

		following := false
		expected := message.CommentsResponse{
			Comments: []message.CommentResponse{
				cm2.ResponseComment(following),
				cm1.ResponseComment(following),
			},
		}

		apiUrl := fmt.Sprintf("/api/v1/articles/%d/comments", barArticle.ID)
		req := httptest.NewRequest(http.MethodGet, apiUrl, nil)

		w := httptest.NewRecorder()
		c, _ := ctxWithToken(t, w, req, fooUser.ID, time.Now().Add(-time.Hour))
		c.AddParam("slug", strconv.Itoa(int(barArticle.ID)))

		h.GetComments(c)

		var actual message.CommentsResponse
		err := json.NewDecoder(w.Result().Body).Decode(&actual)
		if err != nil {
			t.Fatal(err)
		}
		defer w.Result().Body.Close()

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.ElementsMatch(t, expected.Comments, actual.Comments)
	})

	t.Run("DeleteComment", func(t *testing.T) {
		fooUser := createRandomUser(t, lct.DB())
		barUser := createRandomUser(t, lct.DB())

		barArticle := createRandomArticle(
			t, lct.DB(),
			barUser.ID,
			[]string{test.RandomString(t, 10), test.RandomString(t, 10)},
		)

		cm := createRandomComment(t, lct.DB(), barArticle.ID, fooUser.ID)

		apiUrl := fmt.Sprintf("/api/v1/articles/%d/comments/%d", barArticle.ID, cm.ID)
		req := httptest.NewRequest(http.MethodDelete, apiUrl, nil)

		w := httptest.NewRecorder()
		c, _ := ctxWithToken(t, w, req, fooUser.ID, time.Now().Add(-time.Hour))
		c.AddParam("slug", strconv.Itoa(int(barArticle.ID)))
		c.AddParam("id", strconv.Itoa(int(cm.ID)))

		h.DeleteComment(c)

		cms, err := h.as.GetComments(context.Background(), barArticle)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Empty(t, cms)
	})
}
