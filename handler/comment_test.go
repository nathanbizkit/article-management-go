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

		barArticle := createRandomArticle(t, lct.DB(), barUser.ID)

		randStr := test.RandomString(t, 20)
		comment := model.Comment{
			Body:   randStr,
			Author: *fooUser,
		}
		expected := comment.ResponseComment(false)

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
		ctx, _ := ctxWithToken(t, lct.Environ(), w, req, fooUser.ID, time.Now())
		ctx.AddParam("slug", strconv.Itoa(int(barArticle.ID)))

		h.CreateComment(ctx)

		var actual message.CommentResponse
		err = json.NewDecoder(w.Result().Body).Decode(&actual)
		if err != nil {
			t.Fatal(err)
		}
		defer w.Result().Body.Close()

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.NotEmpty(t, actual.ID)
		assert.Equal(t, expected.Body, actual.Body)
		assert.Equal(t, expected.Author, actual.Author)
		assert.NotEmpty(t, actual.CreatedAt)
		assert.NotEmpty(t, actual.UpdatedAt)
		assert.Equal(t, actual.CreatedAt, actual.UpdatedAt)
	})

	t.Run("GetComments", func(t *testing.T) {
		fooUser := createRandomUser(t, lct.DB())
		barUser := createRandomUser(t, lct.DB())

		barArticle := createRandomArticle(t, lct.DB(), barUser.ID)

		comment1 := createRandomComment(t, lct.DB(), barArticle.ID, fooUser.ID)
		time.Sleep(1 * time.Second)
		comment2 := createRandomComment(t, lct.DB(), barArticle.ID, barUser.ID)

		expected := message.CommentsResponse{
			Comments: []message.CommentResponse{
				comment2.ResponseComment(false),
				comment1.ResponseComment(false),
			},
		}

		apiUrl := fmt.Sprintf("/api/v1/articles/%d/comments", barArticle.ID)
		req := httptest.NewRequest(http.MethodGet, apiUrl, nil)

		w := httptest.NewRecorder()
		ctx, _ := ctxWithToken(t, lct.Environ(), w, req, fooUser.ID, time.Now())
		ctx.AddParam("slug", strconv.Itoa(int(barArticle.ID)))

		h.GetComments(ctx)

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

		barArticle := createRandomArticle(t, lct.DB(), barUser.ID)

		cm := createRandomComment(t, lct.DB(), barArticle.ID, fooUser.ID)

		apiUrl := fmt.Sprintf("/api/v1/articles/%d/comments/%d", barArticle.ID, cm.ID)
		req := httptest.NewRequest(http.MethodDelete, apiUrl, nil)

		w := httptest.NewRecorder()
		ctx, _ := ctxWithToken(t, lct.Environ(), w, req, fooUser.ID, time.Now())
		ctx.AddParam("slug", strconv.Itoa(int(barArticle.ID)))
		ctx.AddParam("id", strconv.Itoa(int(cm.ID)))

		h.DeleteComment(ctx)

		actualComments, err := h.as.GetComments(context.Background(), barArticle)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Empty(t, actualComments)
	})
}
