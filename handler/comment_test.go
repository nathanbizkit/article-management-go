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
	"github.com/nathanbizkit/article-management-go/message"
	"github.com/nathanbizkit/article-management-go/model"
	"github.com/nathanbizkit/article-management-go/test"
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

		tests := []struct {
			title              string
			reqUser            *model.User
			reqSlug            string
			reqBody            *message.CreateCommentRequest
			expectedStatusCode int
			expectedBody       message.CommentResponse
			expectedError      map[string]interface{}
			hasError           bool
		}{
			{
				"create comment: success",
				fooUser,
				strconv.Itoa(int(barArticle.ID)),
				&message.CreateCommentRequest{
					Body: randStr,
				},
				http.StatusOK,
				comment.ResponseComment(false),
				nil,
				false,
			},
			{
				"create comment: wrong current user id",
				&model.User{ID: 0},
				strconv.Itoa(int(barArticle.ID)),
				&message.CreateCommentRequest{},
				http.StatusNotFound,
				message.CommentResponse{},
				map[string]interface{}{"error": "current user not found"},
				true,
			},
			{
				"create comment: invalid slug",
				fooUser,
				"invalid_slug",
				&message.CreateCommentRequest{},
				http.StatusBadRequest,
				message.CommentResponse{},
				map[string]interface{}{"error": "invalid slug"},
				true,
			},
			{
				"create comment: wrong slug",
				fooUser,
				"0",
				&message.CreateCommentRequest{},
				http.StatusNotFound,
				message.CommentResponse{},
				map[string]interface{}{"error": "article not found"},
				true,
			},
			{
				"create comment: no body",
				fooUser,
				strconv.Itoa(int(barArticle.ID)),
				&message.CreateCommentRequest{
					Body: "",
				},
				http.StatusBadRequest,
				message.CommentResponse{},
				map[string]interface{}{"error": "validation error: Body: cannot be blank."},
				true,
			},
		}

		for _, tt := range tests {
			body, err := json.Marshal(tt.reqBody)
			if err != nil {
				t.Fatal(err)
			}

			apiUrl := fmt.Sprintf("/api/v1/articles/%v/comments", tt.reqSlug)
			req := httptest.NewRequest(http.MethodPost, apiUrl, bytes.NewReader(body))

			w := httptest.NewRecorder()
			ctx, _ := ctxWithToken(t, lct.Environ(), w, req, tt.reqUser.ID, time.Now())
			ctx.AddParam("slug", tt.reqSlug)

			h.CreateComment(ctx)

			assert.Equal(t, tt.expectedStatusCode, w.Result().StatusCode, tt.title)

			if tt.hasError {
				actualBody := test.GetResponseBody[map[string]interface{}](t, w.Result())
				assert.Equal(t, tt.expectedError, actualBody, tt.title)
			} else {
				actualBody := test.GetResponseBody[message.CommentResponse](t, w.Result())
				assert.NotEmpty(t, actualBody.ID, tt.title)
				assert.Equal(t, tt.expectedBody.Body, actualBody.Body, tt.title)
				assert.Equal(t, tt.expectedBody.Author, actualBody.Author, tt.title)
				assert.NotEmpty(t, actualBody.CreatedAt, tt.title)
				assert.NotEmpty(t, actualBody.UpdatedAt, tt.title)
				assert.Equal(t, actualBody.CreatedAt, actualBody.UpdatedAt, tt.title)
			}
		}
	})

	t.Run("GetComments", func(t *testing.T) {
		fooUser := createRandomUser(t, lct.DB())
		barUser := createRandomUser(t, lct.DB())

		fooArticle := createRandomArticle(t, lct.DB(), fooUser.ID)
		barArticle := createRandomArticle(t, lct.DB(), barUser.ID)

		comment1 := createRandomComment(t, lct.DB(), barArticle.ID, fooUser.ID)
		time.Sleep(1 * time.Second)
		comment2 := createRandomComment(t, lct.DB(), barArticle.ID, barUser.ID)

		tests := []struct {
			title              string
			reqUser            *model.User
			reqSlug            string
			expectedStatusCode int
			expectedBody       message.CommentsResponse
			expectedError      map[string]interface{}
			hasError           bool
		}{
			{
				"get comments from barArticle: success",
				fooUser,
				strconv.Itoa(int(barArticle.ID)),
				http.StatusOK,
				message.CommentsResponse{
					Comments: []message.CommentResponse{
						comment2.ResponseComment(false),
						comment1.ResponseComment(false),
					},
				},
				nil,
				false,
			},
			{
				"get comments from barArticle: success (empty comments)",
				fooUser,
				strconv.Itoa(int(fooArticle.ID)),
				http.StatusOK,
				message.CommentsResponse{
					Comments: []message.CommentResponse{},
				},
				nil,
				false,
			},
			{
				"get comments from barArticle (without auth): success",
				&model.User{ID: 0},
				strconv.Itoa(int(barArticle.ID)),
				http.StatusOK,
				message.CommentsResponse{
					Comments: []message.CommentResponse{
						comment2.ResponseComment(false),
						comment1.ResponseComment(false),
					},
				},
				nil,
				false,
			},
			{
				"get comments from barArticle (without auth): success (empty comments)",
				&model.User{ID: 0},
				strconv.Itoa(int(fooArticle.ID)),
				http.StatusOK,
				message.CommentsResponse{
					Comments: []message.CommentResponse{},
				},
				nil,
				false,
			},
			{
				"get comments: invalid slug",
				fooUser,
				"invalid_slug",
				http.StatusBadRequest,
				message.CommentsResponse{},
				map[string]interface{}{"error": "invalid slug"},
				true,
			},
			{
				"get comments: wrong slug",
				fooUser,
				"0",
				http.StatusNotFound,
				message.CommentsResponse{},
				map[string]interface{}{"error": "article not found"},
				true,
			},
		}

		for _, tt := range tests {
			apiUrl := fmt.Sprintf("/api/v1/articles/%v/comments", tt.reqSlug)
			req := httptest.NewRequest(http.MethodGet, apiUrl, nil)

			w := httptest.NewRecorder()
			ctx, _ := ctxWithToken(t, lct.Environ(), w, req, tt.reqUser.ID, time.Now())
			ctx.AddParam("slug", tt.reqSlug)

			h.GetComments(ctx)

			assert.Equal(t, tt.expectedStatusCode, w.Result().StatusCode, tt.title)

			if tt.hasError {
				actualBody := test.GetResponseBody[map[string]interface{}](t, w.Result())
				assert.Equal(t, tt.expectedError, actualBody, tt.title)
			} else {
				actualBody := test.GetResponseBody[message.CommentsResponse](t, w.Result())
				assert.Equal(t, tt.expectedBody, actualBody, tt.title)
			}
		}
	})

	t.Run("DeleteComment", func(t *testing.T) {
		fooUser := createRandomUser(t, lct.DB())
		barUser := createRandomUser(t, lct.DB())
		bazUser := createRandomUser(t, lct.DB())

		fooArticle := createRandomArticle(t, lct.DB(), fooUser.ID)

		barArticle := createRandomArticle(t, lct.DB(), barUser.ID)
		barComment := createRandomComment(t, lct.DB(), barArticle.ID, fooUser.ID)

		bazArticle := createRandomArticle(t, lct.DB(), bazUser.ID)
		bazComment := createRandomComment(t, lct.DB(), bazArticle.ID, bazUser.ID)

		tests := []struct {
			title              string
			reqUser            *model.User
			reqSlug            string
			reqID              string
			expectedStatusCode int
			expectedError      map[string]interface{}
			hasError           bool
		}{
			{
				"delete comment: success",
				fooUser,
				strconv.Itoa(int(barArticle.ID)),
				strconv.Itoa(int(barComment.ID)),
				http.StatusNoContent,
				nil,
				false,
			},
			{
				"delete comment: wrong current user id",
				&model.User{ID: 0},
				strconv.Itoa(int(barArticle.ID)),
				strconv.Itoa(int(barComment.ID)),
				http.StatusNotFound,
				map[string]interface{}{"error": "current user not found"},
				true,
			},
			{
				"delete comment: invalid slug",
				fooUser,
				"invalid_slug",
				strconv.Itoa(int(barComment.ID)),
				http.StatusBadRequest,
				map[string]interface{}{"error": "invalid slug"},
				true,
			},
			{
				"delete comment: invalid comment id",
				fooUser,
				strconv.Itoa(int(barArticle.ID)),
				"invalid_id",
				http.StatusBadRequest,
				map[string]interface{}{"error": "invalid comment id"},
				true,
			},
			{
				"delete comment: wrong comment id",
				fooUser,
				strconv.Itoa(int(barArticle.ID)),
				"0",
				http.StatusNotFound,
				map[string]interface{}{"error": "comment not found"},
				true,
			},
			{
				"delete comment: comment is not from the article",
				fooUser,
				strconv.Itoa(int(fooArticle.ID)),
				strconv.Itoa(int(bazComment.ID)),
				http.StatusBadRequest,
				map[string]interface{}{"error": "the comment is not from this article"},
				true,
			},
			{
				"delete comment: forbidden to delete other user's comment",
				fooUser,
				strconv.Itoa(int(bazArticle.ID)),
				strconv.Itoa(int(bazComment.ID)),
				http.StatusForbidden,
				map[string]interface{}{"error": "forbidden"},
				true,
			},
		}

		for _, tt := range tests {
			apiUrl := fmt.Sprintf("/api/v1/articles/%v/comments/%v", tt.reqSlug, tt.reqID)
			req := httptest.NewRequest(http.MethodDelete, apiUrl, nil)

			w := httptest.NewRecorder()
			ctx, _ := ctxWithToken(t, lct.Environ(), w, req, tt.reqUser.ID, time.Now())
			ctx.AddParam("slug", tt.reqSlug)
			ctx.AddParam("id", tt.reqID)

			h.DeleteComment(ctx)

			assert.Equal(t, tt.expectedStatusCode, w.Result().StatusCode, tt.title)

			if tt.hasError {
				actualBody := test.GetResponseBody[map[string]interface{}](t, w.Result())
				assert.Equal(t, tt.expectedError, actualBody, tt.title)
			} else {
				reqID, err := strconv.ParseUint(tt.reqID, 10, 0)
				if err != nil {
					t.Fatal(err)
				}

				actualComment, err := h.as.GetCommentByID(context.Background(), uint(reqID))

				assert.Error(t, err, tt.title)
				assert.Nil(t, actualComment, tt.title)
			}
		}
	})
}
