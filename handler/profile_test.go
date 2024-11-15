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
	h, lct := setup(t)

	fooUser := createRandomUser(t, lct.DB())
	barUser := createRandomUser(t, lct.DB())

	t.Run("ShowProfile", func(t *testing.T) {
		following := false
		expected := barUser.ResponseProfile(following)

		apiUrl := fmt.Sprintf("/api/profiles/%s", barUser.Username)
		req := httptest.NewRequest(http.MethodGet, apiUrl, nil)

		w := httptest.NewRecorder()
		c, _ := ctxWithToken(t, w, req, fooUser.ID, time.Now().Add(-time.Hour))
		c.AddParam("username", barUser.Username)

		h.ShowProfile(c)

		var actual message.ProfileResponse
		err := json.NewDecoder(w.Result().Body).Decode(&actual)
		if err != nil {
			t.Fatal(err)
		}
		defer w.Result().Body.Close()

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Equal(t, expected, actual)
	})

	t.Run("FollowUser", func(t *testing.T) {
		err := h.us.Unfollow(context.Background(), fooUser, barUser)
		if err != nil {
			t.Fatal(err)
		}

		following := true
		expected := barUser.ResponseProfile(following)

		apiUrl := fmt.Sprintf("/api/profiles/%s/follow", barUser.Username)
		req := httptest.NewRequest(http.MethodPost, apiUrl, nil)

		w := httptest.NewRecorder()
		c, _ := ctxWithToken(t, w, req, fooUser.ID, time.Now().Add(-time.Hour))
		c.AddParam("username", barUser.Username)

		h.FollowUser(c)

		var actual message.ProfileResponse
		err = json.NewDecoder(w.Result().Body).Decode(&actual)
		if err != nil {
			t.Fatal(err)
		}
		defer w.Result().Body.Close()

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Equal(t, expected, actual)
	})

	t.Run("UnfollowUser", func(t *testing.T) {
		err := h.us.Follow(context.Background(), fooUser, barUser)
		if err != nil {
			t.Fatal(err)
		}

		following := false
		expected := barUser.ResponseProfile(following)

		apiUrl := fmt.Sprintf("/api/profiles/%s/follow", barUser.Username)
		req := httptest.NewRequest(http.MethodDelete, apiUrl, nil)

		w := httptest.NewRecorder()
		c, _ := ctxWithToken(t, w, req, fooUser.ID, time.Now().Add(-time.Hour))
		c.AddParam("username", barUser.Username)

		h.UnfollowUser(c)

		var actual message.ProfileResponse
		err = json.NewDecoder(w.Result().Body).Decode(&actual)
		if err != nil {
			t.Fatal(err)
		}
		defer w.Result().Body.Close()

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Equal(t, expected, actual)
	})
}
