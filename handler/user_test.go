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

		tests := []struct {
			title              string
			reqBody            *message.LoginUserRequest
			expectedStatusCode int
			expectedUserID     uint
			hasError           bool
		}{
			{
				"login to fooUser: success",
				&message.LoginUserRequest{
					Email:    fooUser.Email,
					Password: userPassword,
				},
				http.StatusOK,
				fooUser.ID,
				false,
			},
			{
				"login to fooUser: wrong email",
				&message.LoginUserRequest{
					Email:    "fooooo@example.com",
					Password: userPassword,
				},
				http.StatusNotFound,
				uint(0),
				true,
			},
			{
				"login to fooUser: wrong password",
				&message.LoginUserRequest{
					Email:    fooUser.Email,
					Password: "wrong_password",
				},
				http.StatusBadRequest,
				uint(0),
				true,
			},
		}

		for _, tt := range tests {
			body, err := json.Marshal(tt.reqBody)
			if err != nil {
				t.Fatal(err)
			}

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/login", bytes.NewReader(body))

			h.Login(ctx)

			actualCookies := w.Result().Header.Values("Set-Cookie")

			assert.Equal(t, tt.expectedStatusCode, w.Result().StatusCode, tt.title)

			if tt.hasError {
				assert.Empty(t, actualCookies, tt.title)
			} else {
				assert.Len(t, actualCookies, 2)

				for _, actualCookie := range actualCookies {
					cookie, err := http.ParseSetCookie(actualCookie)
					if err != nil {
						t.Fatal(err)
					}

					assert.NotEmpty(t, cookie.Value)
				}
			}
		}
	})

	t.Run("Register", func(t *testing.T) {
		fooUser := createRandomUser(t, lct.DB())

		randStr := test.RandomString(t, 10)
		barUser := model.User{
			Username: randStr,
			Email:    fmt.Sprintf("%s@example.com", randStr),
			Password: userPassword,
			Name:     strings.ToUpper(randStr),
		}

		tests := []struct {
			title              string
			reqBody            *message.CreateUserRequest
			expectedStatusCode int
			expected           message.ProfileResponse
			hasError           bool
		}{
			{
				"register barUser: success",
				&message.CreateUserRequest{
					Username: barUser.Username,
					Email:    barUser.Email,
					Password: barUser.Password,
					Name:     barUser.Name,
				},
				http.StatusOK,
				barUser.ResponseProfile(false),
				false,
			},
			{
				"register barUser: no username",
				&message.CreateUserRequest{
					Username: "",
					Email:    barUser.Email,
					Password: barUser.Password,
					Name:     barUser.Name,
				},
				http.StatusBadRequest,
				message.ProfileResponse{},
				true,
			},
			{
				"register barUser: no password",
				&message.CreateUserRequest{
					Username: barUser.Username,
					Email:    barUser.Email,
					Password: "",
					Name:     barUser.Name,
				},
				http.StatusBadRequest,
				message.ProfileResponse{},
				true,
			},
			{
				"register barUser: no email",
				&message.CreateUserRequest{
					Username: barUser.Username,
					Email:    "",
					Password: barUser.Password,
					Name:     barUser.Name,
				},
				http.StatusBadRequest,
				message.ProfileResponse{},
				true,
			},
			{
				"register barUser: no name",
				&message.CreateUserRequest{
					Username: barUser.Username,
					Email:    barUser.Email,
					Password: barUser.Password,
					Name:     "",
				},
				http.StatusBadRequest,
				message.ProfileResponse{},
				true,
			},
			{
				"register barUser: username already exists",
				&message.CreateUserRequest{
					Username: fooUser.Username,
					Email:    barUser.Email,
					Password: barUser.Password,
					Name:     barUser.Name,
				},
				http.StatusInternalServerError,
				message.ProfileResponse{},
				true,
			},
			{
				"register barUser: email already exists",
				&message.CreateUserRequest{
					Username: barUser.Username,
					Email:    fooUser.Email,
					Password: barUser.Password,
					Name:     barUser.Name,
				},
				http.StatusInternalServerError,
				message.ProfileResponse{},
				true,
			},
		}

		for _, tt := range tests {
			body, err := json.Marshal(tt.reqBody)
			if err != nil {
				t.Fatal(err)
			}

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/register", bytes.NewReader(body))

			h.Register(ctx)

			var actual message.ProfileResponse
			err = json.NewDecoder(w.Result().Body).Decode(&actual)
			if err != nil {
				t.Fatal(err)
			}
			defer w.Result().Body.Close()

			actualCookies := w.Result().Header.Values("Set-Cookie")

			assert.Equal(t, tt.expectedStatusCode, w.Result().StatusCode, tt.title)
			assert.Equal(t, tt.expected, actual, tt.title)

			if tt.hasError {
				assert.Empty(t, actualCookies, tt.title)
			} else {
				assert.Len(t, actualCookies, 2)

				for _, actualCookie := range actualCookies {
					cookie, err := http.ParseSetCookie(actualCookie)
					if err != nil {
						t.Fatal(err)
					}

					assert.NotEmpty(t, cookie.Value)
				}
			}
		}
	})

	t.Run("RefreshToken", func(t *testing.T) {
		fooUser := createRandomUser(t, lct.DB())

		req := httptest.NewRequest(http.MethodPost, "/api/v1/refresh_token", nil)
		w := httptest.NewRecorder()
		c, token := ctxWithToken(t, lct.Environ(), w, req, fooUser.ID, time.Now().Add(-time.Hour))

		h.RefreshToken(c)

		actualCookies := w.Result().Header.Values("Set-Cookie")

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Len(t, actualCookies, 2)

		for _, actualCookie := range actualCookies {
			cookie, err := http.ParseSetCookie(actualCookie)
			if err != nil {
				t.Fatal(err)
			}

			assert.NotEmpty(t, cookie.Value)

			switch cookie.Name {
			case "session":
				assert.NotEqual(t, token.Token, cookie.Value)
			case "refreshToken":
				assert.NotEqual(t, token.RefreshToken, cookie.Value)
			}
		}
	})

	t.Run("GetCurrentUser", func(t *testing.T) {
		fooUser := createRandomUser(t, lct.DB())

		expected := fooUser.ResponseProfile(false)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
		w := httptest.NewRecorder()
		c, token := ctxWithToken(t, lct.Environ(), w, req, fooUser.ID, time.Now().Add(-time.Hour))

		h.GetCurrentUser(c)

		var actual message.ProfileResponse
		err := json.NewDecoder(w.Result().Body).Decode(&actual)
		if err != nil {
			t.Fatal(err)
		}
		defer w.Result().Body.Close()

		actualCookies := w.Result().Header.Values("Set-Cookie")

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.Equal(t, expected, actual)
		assert.Len(t, actualCookies, 2)

		for _, actualCookie := range actualCookies {
			cookie, err := http.ParseSetCookie(actualCookie)
			if err != nil {
				t.Fatal(err)
			}

			assert.NotEmpty(t, cookie.Value)

			switch cookie.Name {
			case "session":
				assert.NotEqual(t, token.Token, cookie.Value)
			case "refreshToken":
				assert.NotEqual(t, token.RefreshToken, cookie.Value)
			}
		}
	})

	t.Run("UpdateCurrentUser", func(t *testing.T) {
		fooUser := createRandomUser(t, lct.DB())

		randStr := test.RandomString(t, 15)
		user := model.User{
			ID:        fooUser.ID,
			Username:  randStr,
			Email:     fmt.Sprintf("%s@example.com", randStr),
			Password:  fmt.Sprintf("%s_NEW", userPassword),
			Name:      randStr,
			Bio:       randStr,
			Image:     "https://imgur.com/image.jpg",
			CreatedAt: fooUser.CreatedAt,
			UpdatedAt: fooUser.UpdatedAt,
		}
		expected := user.ResponseProfile(false)

		tests := []struct {
			title              string
			reqBody            *message.UpdateUserRequest
			expectedStatusCode int
			expected           message.ProfileResponse
		}{
			{
				"update fooUser: success",
				&message.UpdateUserRequest{
					Username: user.Username,
					Email:    user.Email,
					Password: user.Password,
					Name:     user.Name,
					Bio:      randStr,
					Image:    "https://imgur.com/image.jpg",
				},
				http.StatusOK,
				expected,
			},
			{
				"update fooUser: ignore zero-value field",
				&message.UpdateUserRequest{
					Username: "",
					Email:    "",
					Password: "",
					Name:     "",
					Bio:      randStr,
					Image:    "https://imgur.com/image.jpg",
				},
				http.StatusOK,
				expected,
			},
		}

		for _, tt := range tests {
			body, err := json.Marshal(tt.reqBody)
			if err != nil {
				t.Fatal(err)
			}

			req := httptest.NewRequest(http.MethodPut, "/api/v1/me", bytes.NewReader(body))
			w := httptest.NewRecorder()
			c, token := ctxWithToken(t, lct.Environ(), w, req, fooUser.ID, time.Now().Add(-time.Hour))

			h.UpdateCurrentUser(c)

			var actual message.ProfileResponse
			err = json.NewDecoder(w.Result().Body).Decode(&actual)
			if err != nil {
				t.Fatal(err)
			}
			defer w.Result().Body.Close()

			actualCookies := w.Result().Header.Values("Set-Cookie")

			assert.Equal(t, http.StatusOK, w.Result().StatusCode, tt.title)
			assert.Equal(t, expected, actual, tt.title)
			assert.Len(t, actualCookies, 2, tt.title)

			for _, actualCookie := range actualCookies {
				cookie, err := http.ParseSetCookie(actualCookie)
				if err != nil {
					t.Fatal(err)
				}

				assert.NotEmpty(t, cookie.Value, tt.title)

				switch cookie.Name {
				case "session":
					assert.NotEqual(t, token.Token, cookie.Value, tt.title)
				case "refreshToken":
					assert.NotEqual(t, token.RefreshToken, cookie.Value, tt.title)
				}
			}
		}
	})
}
