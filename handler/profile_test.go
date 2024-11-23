package handler

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nathanbizkit/article-management/message"
	"github.com/nathanbizkit/article-management/model"
	"github.com/nathanbizkit/article-management/test"
	"github.com/stretchr/testify/assert"
)

func TestIntegration_ProfileHandler(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration tests.")
	}

	gin.SetMode("test")
	h, lct := setup(t)

	t.Run("ShowProfile", func(t *testing.T) {
		fooUser := createRandomUser(t, lct.DB())
		barUser := createRandomUser(t, lct.DB())

		tests := []struct {
			title              string
			reqUser            *model.User
			reqUsername        string
			expectedStatusCode int
			expectedBody       message.ProfileResponse
			expectedError      map[string]interface{}
			hasError           bool
		}{
			{
				"show profile: barUser profile",
				fooUser,
				barUser.Username,
				http.StatusOK,
				barUser.ResponseProfile(false),
				nil,
				false,
			},
			{
				"show profile: wrong current user id",
				&model.User{ID: 0},
				barUser.Username,
				http.StatusNotFound,
				message.ProfileResponse{},
				map[string]interface{}{"error": "current user not found"},
				true,
			},
			{
				"show profile: wrong profile username",
				fooUser,
				"unknown_user",
				http.StatusNotFound,
				message.ProfileResponse{},
				map[string]interface{}{"error": "user not found"},
				true,
			},
		}

		for _, tt := range tests {
			apiUrl := fmt.Sprintf("/api/v1/profiles/%s", tt.reqUsername)
			req := httptest.NewRequest(http.MethodGet, apiUrl, nil)

			w := httptest.NewRecorder()
			ctx, _ := ctxWithToken(t, lct.Environ(), w, req, tt.reqUser.ID, time.Now())
			ctx.AddParam("username", tt.reqUsername)

			h.ShowProfile(ctx)

			assert.Equal(t, tt.expectedStatusCode, w.Result().StatusCode, tt.title)

			if tt.hasError {
				actualBody := test.GetResponseBody[map[string]interface{}](t, w.Result())
				assert.Equal(t, tt.expectedError, actualBody, tt.title)
			} else {
				actualBody := test.GetResponseBody[message.ProfileResponse](t, w.Result())
				assert.Equal(t, tt.expectedBody, actualBody, tt.title)
			}
		}
	})

	t.Run("FollowUser", func(t *testing.T) {
		fooUser := createRandomUser(t, lct.DB())
		barUser := createRandomUser(t, lct.DB())
		bazUser := createRandomUser(t, lct.DB())

		err := h.us.Follow(context.Background(), fooUser, bazUser)
		if err != nil {
			t.Fatal(err)
		}

		tests := []struct {
			title              string
			reqUser            *model.User
			reqUsername        string
			expectedStatusCode int
			expectedBody       message.ProfileResponse
			expectedError      map[string]interface{}
			hasError           bool
		}{
			{
				"follow user: fooUser follows barUser",
				fooUser,
				barUser.Username,
				http.StatusOK,
				barUser.ResponseProfile(true),
				nil,
				false,
			},
			{
				"follow user: wrong current user id",
				&model.User{ID: 0},
				barUser.Username,
				http.StatusNotFound,
				message.ProfileResponse{},
				map[string]interface{}{"error": "current user not found"},
				true,
			},
			{
				"follow user: cannot follow yourself",
				fooUser,
				fooUser.Username,
				http.StatusBadRequest,
				message.ProfileResponse{},
				map[string]interface{}{"error": "cannot follow yourself"},
				true,
			},
			{
				"follow user: wrong username",
				fooUser,
				"unknown_user",
				http.StatusNotFound,
				message.ProfileResponse{},
				map[string]interface{}{"error": "user not found"},
				true,
			},
			{
				"follow user: fooUser is already following bazUser",
				fooUser,
				bazUser.Username,
				http.StatusBadRequest,
				message.ProfileResponse{},
				map[string]interface{}{"error": "you are already following this user"},
				true,
			},
		}

		for _, tt := range tests {
			apiUrl := fmt.Sprintf("/api/v1/profiles/%s/follow", tt.reqUsername)
			req := httptest.NewRequest(http.MethodPost, apiUrl, nil)

			w := httptest.NewRecorder()
			ctx, _ := ctxWithToken(t, lct.Environ(), w, req, tt.reqUser.ID, time.Now())
			ctx.AddParam("username", tt.reqUsername)

			h.FollowUser(ctx)

			assert.Equal(t, tt.expectedStatusCode, w.Result().StatusCode, tt.title)

			if tt.hasError {
				actualBody := test.GetResponseBody[map[string]interface{}](t, w.Result())
				assert.Equal(t, tt.expectedError, actualBody, tt.title)
			} else {
				actualBody := test.GetResponseBody[message.ProfileResponse](t, w.Result())
				assert.Equal(t, tt.expectedBody, actualBody, tt.title)

				actualFollowing, err := h.us.IsFollowing(context.Background(), fooUser, barUser)
				if err != nil {
					t.Fatal(err)
				}

				assert.True(t, actualFollowing, tt.title)
			}
		}
	})

	t.Run("UnfollowUser", func(t *testing.T) {
		fooUser := createRandomUser(t, lct.DB())
		barUser := createRandomUser(t, lct.DB())
		bazUser := createRandomUser(t, lct.DB())

		err := h.us.Follow(context.Background(), fooUser, barUser)
		if err != nil {
			t.Fatal(err)
		}

		tests := []struct {
			title              string
			reqUser            *model.User
			reqUsername        string
			expectedStatusCode int
			expectedBody       message.ProfileResponse
			expectedError      map[string]interface{}
			hasError           bool
		}{
			{
				"unfollow user: fooUser unfollows barUser",
				fooUser,
				barUser.Username,
				http.StatusOK,
				barUser.ResponseProfile(false),
				nil,
				false,
			},
			{
				"unfollow user: wrong current user id",
				&model.User{ID: 0},
				barUser.Username,
				http.StatusNotFound,
				message.ProfileResponse{},
				map[string]interface{}{"error": "current user not found"},
				true,
			},
			{
				"unfollow user: cannot unfollow yourself",
				fooUser,
				fooUser.Username,
				http.StatusBadRequest,
				message.ProfileResponse{},
				map[string]interface{}{"error": "cannot unfollow yourself"},
				true,
			},
			{
				"unfollow user: wrong username",
				fooUser,
				"unknown_user",
				http.StatusNotFound,
				message.ProfileResponse{},
				map[string]interface{}{"error": "user not found"},
				true,
			},
			{
				"unfollow user: fooUser is not following bazUser",
				fooUser,
				bazUser.Username,
				http.StatusBadRequest,
				message.ProfileResponse{},
				map[string]interface{}{"error": "you are not following this user"},
				true,
			},
		}

		for _, tt := range tests {
			apiUrl := fmt.Sprintf("/api/v1/profiles/%s/follow", tt.reqUsername)
			req := httptest.NewRequest(http.MethodDelete, apiUrl, nil)

			w := httptest.NewRecorder()
			ctx, _ := ctxWithToken(t, lct.Environ(), w, req, tt.reqUser.ID, time.Now())
			ctx.AddParam("username", tt.reqUsername)

			h.UnfollowUser(ctx)

			assert.Equal(t, tt.expectedStatusCode, w.Result().StatusCode, tt.title)

			if tt.hasError {
				actualBody := test.GetResponseBody[map[string]interface{}](t, w.Result())
				assert.Equal(t, tt.expectedError, actualBody, tt.title)
			} else {
				actualBody := test.GetResponseBody[message.ProfileResponse](t, w.Result())
				assert.Equal(t, tt.expectedBody, actualBody, tt.title)

				actualFollowing, err := h.us.IsFollowing(context.Background(), fooUser, barUser)
				if err != nil {
					t.Fatal(err)
				}

				assert.False(t, actualFollowing, tt.title)
			}
		}
	})
}
