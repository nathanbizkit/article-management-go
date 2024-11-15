package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nathanbizkit/article-management/message"
	"github.com/nathanbizkit/article-management/model"
	"github.com/nathanbizkit/article-management/test"
	"github.com/stretchr/testify/assert"
)

func TestIntegration_UserHandler(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration tests.")
	}

	gin.SetMode("test")
	h, lct := setup(t)

	t.Run("Login", func(t *testing.T) {
		fooUser := createRandomUser(t, lct.DB())

		r := message.LoginUserRequest{
			Email:    fooUser.Email,
			Password: userPassword,
		}

		body, err := json.Marshal(r)
		if err != nil {
			t.Fatal(err)
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewReader(body))

		h.Login(c)

		assertCtx, _ := gin.CreateTestContext(httptest.NewRecorder())
		assertCtx.Request = &http.Request{
			Header: make(http.Header),
		}
		assertCtx.Request.Header.Set(
			"Cookie",
			strings.Join(w.Result().Header.Values("Set-Cookie"), "; "),
		)

		actualUserID, err := h.auth.GetUserID(assertCtx, false)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.NoError(t, err)
		assert.Equal(t, fooUser.ID, actualUserID)
	})

	t.Run("Register", func(t *testing.T) {
		randStr := test.RandomString(t, 10)
		u := model.User{
			Username: randStr,
			Email:    fmt.Sprintf("%s@example.com", randStr),
			Password: userPassword,
			Name:     strings.ToUpper(randStr),
		}

		following := false
		expected := u.ResponseProfile(following)

		r := message.CreateUserRequest{
			Username: u.Username,
			Email:    u.Email,
			Password: u.Password,
			Name:     u.Name,
		}

		body, err := json.Marshal(r)
		if err != nil {
			t.Fatal(err)
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewReader(body))

		h.Register(c)

		var actual message.ProfileResponse
		err = json.NewDecoder(w.Result().Body).Decode(&actual)
		if err != nil {
			t.Fatal(err)
		}
		defer w.Result().Body.Close()

		assertCtx, _ := gin.CreateTestContext(httptest.NewRecorder())
		assertCtx.Request = &http.Request{
			Header: make(http.Header),
		}
		assertCtx.Request.Header.Set(
			"Cookie",
			strings.Join(w.Result().Header.Values("Set-Cookie"), "; "),
		)

		actualUserID, err := h.auth.GetUserID(assertCtx, false)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Equal(t, expected, actual)
		assert.NoError(t, err)
		assert.Greater(t, actualUserID, uint(1))
	})

	t.Run("RefreshToken", func(t *testing.T) {
		fooUser := createRandomUser(t, lct.DB())

		req := httptest.NewRequest(http.MethodPost, "/api/refresh_token", nil)
		w := httptest.NewRecorder()
		c, token := ctxWithToken(t, w, req, fooUser.ID, time.Now().Add(-time.Hour))

		h.RefreshToken(c)

		assertCtx, _ := gin.CreateTestContext(httptest.NewRecorder())
		assertCtx.Request = &http.Request{
			Header: make(http.Header),
		}
		assertCtx.Request.Header.Set(
			"Cookie",
			strings.Join(w.Result().Header.Values("Set-Cookie"), "; "),
		)

		actualToken, err := h.auth.GetCookieToken(assertCtx)
		if err != nil {
			t.Fatal(err)
		}

		actualUserID, err := h.auth.GetUserID(assertCtx, true)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.NotEqual(t, token.Token, actualToken.Token)
		assert.NotEqual(t, token.RefreshToken, actualToken.RefreshToken)
		assert.NoError(t, err)
		assert.Equal(t, fooUser.ID, actualUserID)
	})

	t.Run("GetCurrentUser", func(t *testing.T) {
		fooUser := createRandomUser(t, lct.DB())

		following := false
		expected := fooUser.ResponseProfile(following)

		req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
		w := httptest.NewRecorder()
		c, token := ctxWithToken(t, w, req, fooUser.ID, time.Now().Add(-time.Hour))

		h.GetCurrentUser(c)

		var actual message.ProfileResponse
		err := json.NewDecoder(w.Result().Body).Decode(&actual)
		if err != nil {
			t.Fatal(err)
		}
		defer w.Result().Body.Close()

		assertCtx, _ := gin.CreateTestContext(httptest.NewRecorder())
		assertCtx.Request = &http.Request{
			Header: make(http.Header),
		}
		assertCtx.Request.Header.Set(
			"Cookie",
			strings.Join(w.Result().Header.Values("Set-Cookie"), "; "),
		)

		actualToken, err := h.auth.GetCookieToken(assertCtx)
		if err != nil {
			t.Fatal(err)
		}

		actualUserID, err := h.auth.GetUserID(assertCtx, false)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Equal(t, expected, actual)
		assert.NotEqual(t, token.Token, actualToken.Token)
		assert.NotEqual(t, token.RefreshToken, actualToken.RefreshToken)
		assert.NoError(t, err)
		assert.Equal(t, fooUser.ID, actualUserID)
	})

	t.Run("UpdateCurrentUser", func(t *testing.T) {
		fooUser := createRandomUser(t, lct.DB())

		randStr := test.RandomString(t, 15)
		u := model.User{
			Username: randStr,
			Email:    fmt.Sprintf("%s@example.com", randStr),
			Password: fmt.Sprintf("%s_NEW", userPassword),
			Name:     randStr,
			Bio:      randStr,
			Image:    "https://imgur.com/image.jpg",
		}

		following := false
		expected := u.ResponseProfile(following)

		r := message.UpdateUserRequest{
			Username: u.Username,
			Email:    u.Email,
			Password: u.Password,
			Name:     u.Name,
			Bio:      randStr,
			Image:    "https://imgur.com/image.jpg",
		}

		body, err := json.Marshal(r)
		if err != nil {
			t.Fatal(err)
		}

		req := httptest.NewRequest(http.MethodPut, "/api/me", bytes.NewReader(body))
		w := httptest.NewRecorder()
		c, token := ctxWithToken(t, w, req, fooUser.ID, time.Now().Add(-time.Hour))

		h.UpdateCurrentUser(c)

		var actual message.ProfileResponse
		err = json.NewDecoder(w.Result().Body).Decode(&actual)
		if err != nil {
			t.Fatal(err)
		}
		defer w.Result().Body.Close()

		assertCtx, _ := gin.CreateTestContext(httptest.NewRecorder())
		assertCtx.Request = &http.Request{
			Header: make(http.Header),
		}
		assertCtx.Request.Header.Set(
			"Cookie",
			strings.Join(w.Result().Header.Values("Set-Cookie"), "; "),
		)

		actualToken, err := h.auth.GetCookieToken(assertCtx)
		if err != nil {
			t.Fatal(err)
		}

		actualUserID, err := h.auth.GetUserID(assertCtx, false)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Equal(t, expected, actual)
		assert.NotEqual(t, token.Token, actualToken.Token)
		assert.NotEqual(t, token.RefreshToken, actualToken.RefreshToken)
		assert.NoError(t, err)
		assert.Equal(t, fooUser.ID, actualUserID)
	})
}
