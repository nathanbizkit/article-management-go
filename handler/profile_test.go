package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nathanbizkit/article-management/message"
	"github.com/stretchr/testify/assert"
)

func TestIntegration_ProfileHandler(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration tests.")
	}

	gin.SetMode("test")

	h, lct := setUp(t)

	fooUser := createUser(t, lct.DB())
	barUser := createUser(t, lct.DB())

	t.Run("ShowProfile", func(t *testing.T) {
		following := false
		expected := barUser.ResponseProfile(following)

		tokenTime := time.Now().Add(-5 * time.Hour)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet,
			fmt.Sprintf("/api/profiles/%s", barUser.Username), nil)
		c, _ := ctxWithToken(t, w, req, fooUser.ID, tokenTime)
		c.AddParam("username", barUser.Username)

		h.ShowProfile(c)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)

		var actual message.ProfileResponse
		err := json.NewDecoder(w.Result().Body).Decode(&actual)
		if err != nil {
			t.Fatal(err)
		}
		w.Result().Body.Close()

		assert.Equal(t, expected, actual)
	})

	t.Run("FollowUser", func(t *testing.T) {
		err := h.us.Unfollow(context.Background(), fooUser, barUser)
		if err != nil {
			t.Fatal(err)
		}

		following := true
		expected := barUser.ResponseProfile(following)

		tokenTime := time.Now().Add(-5 * time.Hour)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet,
			fmt.Sprintf("/api/profiles/%s/follow", barUser.Username), nil)
		c, _ := ctxWithToken(t, w, req, fooUser.ID, tokenTime)
		c.AddParam("username", barUser.Username)

		h.FollowUser(c)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)

		var actual message.ProfileResponse
		err = json.NewDecoder(w.Result().Body).Decode(&actual)
		if err != nil {
			t.Fatal(err)
		}
		w.Result().Body.Close()

		assert.Equal(t, expected, actual)
	})

	t.Run("UnfollowUser", func(t *testing.T) {
		err := h.us.Follow(context.Background(), fooUser, barUser)
		if err != nil {
			t.Fatal(err)
		}

		following := false
		expected := barUser.ResponseProfile(following)

		tokenTime := time.Now().Add(-5 * time.Hour)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet,
			fmt.Sprintf("/api/profiles/%s/unfollow", barUser.Username), nil)
		c, _ := ctxWithToken(t, w, req, fooUser.ID, tokenTime)
		c.AddParam("username", barUser.Username)

		h.UnfollowUser(c)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)

		var actual message.ProfileResponse
		err = json.NewDecoder(w.Result().Body).Decode(&actual)
		if err != nil {
			t.Fatal(err)
		}
		w.Result().Body.Close()

		assert.Equal(t, expected, actual)
	})
}
