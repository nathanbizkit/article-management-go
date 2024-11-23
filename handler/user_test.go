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
			expectedError      map[string]interface{}
			hasError           bool
		}{
			{
				"login: success",
				&message.LoginUserRequest{
					Email:    fooUser.Email,
					Password: userPassword,
				},
				http.StatusNoContent,
				nil,
				false,
			},
			{
				"login: wrong email",
				&message.LoginUserRequest{
					Email:    "fooooo@example.com",
					Password: userPassword,
				},
				http.StatusNotFound,
				map[string]interface{}{"error": "user not found"},
				true,
			},
			{
				"login: wrong password",
				&message.LoginUserRequest{
					Email:    fooUser.Email,
					Password: "wrong_password",
				},
				http.StatusBadRequest,
				map[string]interface{}{"error": "invalid password"},
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
				actualBody := test.GetResponseBody[map[string]interface{}](t, w.Result())
				assert.Equal(t, tt.expectedError, actualBody, tt.title)
				assert.Empty(t, actualCookies, tt.title)
			} else {
				assert.Len(t, actualCookies, 2, tt.title)

				for _, actualCookie := range actualCookies {
					cookie, err := http.ParseSetCookie(actualCookie)
					if err != nil {
						t.Fatal(err)
					}
					assert.NotEmpty(t, cookie.Value, tt.title)
				}
			}
		}
	})

	t.Run("Register", func(t *testing.T) {
		shortMaxLenString := strings.Repeat("a", 101)
		passwordMaxLenString := strings.Repeat("a", 51)

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
			expectedBody       message.ProfileResponse
			expectedError      map[string]interface{}
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
				nil,
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
				map[string]interface{}{"error": "validation error: Username: cannot be blank."},
				true,
			},
			{
				"register barUser: invalid username format",
				&message.CreateUserRequest{
					Username: "_invalid@@username_",
					Email:    barUser.Email,
					Password: barUser.Password,
					Name:     barUser.Name,
				},
				http.StatusBadRequest,
				message.ProfileResponse{},
				map[string]interface{}{"error": "validation error: Username: must be in a valid format."},
				true,
			},
			{
				"register barUser: username is too short",
				&message.CreateUserRequest{
					Username: "abc",
					Email:    barUser.Email,
					Password: barUser.Password,
					Name:     barUser.Name,
				},
				http.StatusBadRequest,
				message.ProfileResponse{},
				map[string]interface{}{"error": "validation error: Username: the length must be between 5 and 100."},
				true,
			},
			{
				"register barUser: username is too long",
				&message.CreateUserRequest{
					Username: shortMaxLenString,
					Email:    barUser.Email,
					Password: barUser.Password,
					Name:     barUser.Name,
				},
				http.StatusBadRequest,
				message.ProfileResponse{},
				map[string]interface{}{"error": "validation error: Username: the length must be between 5 and 100."},
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
				map[string]interface{}{"error": "validation error: Password: cannot be blank."},
				true,
			},
			{
				"register barUser: password is too short",
				&message.CreateUserRequest{
					Username: barUser.Username,
					Email:    barUser.Email,
					Password: "abc",
					Name:     barUser.Name,
				},
				http.StatusBadRequest,
				message.ProfileResponse{},
				map[string]interface{}{"error": "validation error: Password: the length must be between 7 and 50."},
				true,
			},
			{
				"register barUser: password is too long",
				&message.CreateUserRequest{
					Username: barUser.Username,
					Email:    barUser.Email,
					Password: passwordMaxLenString,
					Name:     barUser.Name,
				},
				http.StatusBadRequest,
				message.ProfileResponse{},
				map[string]interface{}{"error": "validation error: Password: the length must be between 7 and 50."},
				true,
			},
			{
				"register barUser: weak password",
				&message.CreateUserRequest{
					Username: barUser.Username,
					Email:    barUser.Email,
					Password: "password",
					Name:     barUser.Name,
				},
				http.StatusBadRequest,
				message.ProfileResponse{},
				map[string]interface{}{"error": "validation error: Password: must have at least one uppercase, one number, one symbol or punctuation."},
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
				map[string]interface{}{"error": "validation error: Email: cannot be blank."},
				true,
			},
			{
				"register barUser: email is too short",
				&message.CreateUserRequest{
					Username: barUser.Username,
					Email:    "abc",
					Password: barUser.Password,
					Name:     barUser.Name,
				},
				http.StatusBadRequest,
				message.ProfileResponse{},
				map[string]interface{}{"error": "validation error: Email: the length must be between 5 and 100."},
				true,
			},
			{
				"register barUser: email is too long",
				&message.CreateUserRequest{
					Username: barUser.Username,
					Email:    shortMaxLenString,
					Password: barUser.Password,
					Name:     barUser.Name,
				},
				http.StatusBadRequest,
				message.ProfileResponse{},
				map[string]interface{}{"error": "validation error: Email: the length must be between 5 and 100."},
				true,
			},
			{
				"register barUser: invalid email format",
				&message.CreateUserRequest{
					Username: barUser.Username,
					Email:    "invalid_email_format",
					Password: barUser.Password,
					Name:     barUser.Name,
				},
				http.StatusBadRequest,
				message.ProfileResponse{},
				map[string]interface{}{"error": "validation error: Email: must be a valid email address."},
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
				map[string]interface{}{"error": "validation error: Name: cannot be blank."},
				true,
			},
			{
				"register barUser: name is too short",
				&message.CreateUserRequest{
					Username: barUser.Username,
					Email:    barUser.Email,
					Password: barUser.Password,
					Name:     "abc",
				},
				http.StatusBadRequest,
				message.ProfileResponse{},
				map[string]interface{}{"error": "validation error: Name: the length must be between 5 and 100."},
				true,
			},
			{
				"register barUser: name is too long",
				&message.CreateUserRequest{
					Username: barUser.Username,
					Email:    barUser.Email,
					Password: barUser.Password,
					Name:     shortMaxLenString,
				},
				http.StatusBadRequest,
				message.ProfileResponse{},
				map[string]interface{}{"error": "validation error: Name: the length must be between 5 and 100."},
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
				map[string]interface{}{"error": "failed to create user"},
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
				map[string]interface{}{"error": "failed to create user"},
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

			actualCookies := w.Result().Header.Values("Set-Cookie")

			assert.Equal(t, tt.expectedStatusCode, w.Result().StatusCode, tt.title)

			if tt.hasError {
				actualBody := test.GetResponseBody[map[string]interface{}](t, w.Result())
				assert.Equal(t, tt.expectedError, actualBody, tt.title)
				assert.Empty(t, actualCookies, tt.title)
			} else {
				actualBody := test.GetResponseBody[message.ProfileResponse](t, w.Result())
				assert.Equal(t, tt.expectedBody, actualBody, tt.title)
				assert.Len(t, actualCookies, 2, tt.title)

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

		tests := []struct {
			title              string
			reqUser            *model.User
			expectedStatusCode int
			expectedError      map[string]interface{}
			hasError           bool
		}{
			{
				"refresh token: success",
				fooUser,
				http.StatusNoContent,
				nil,
				false,
			},
			{
				"refresh token: wrong user id",
				&model.User{ID: 0},
				http.StatusNotFound,
				map[string]interface{}{"error": "user not found"},
				true,
			},
		}

		for _, tt := range tests {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/refresh_token", nil)
			w := httptest.NewRecorder()
			c, token := ctxWithToken(t, lct.Environ(), w, req, tt.reqUser.ID, time.Now().Add(-time.Hour))

			h.RefreshToken(c)

			actualCookies := w.Result().Header.Values("Set-Cookie")

			assert.Equal(t, tt.expectedStatusCode, w.Result().StatusCode, tt.title)

			if tt.hasError {
				actualBody := test.GetResponseBody[map[string]interface{}](t, w.Result())
				assert.Equal(t, tt.expectedError, actualBody, tt.title)
				assert.Empty(t, actualCookies, tt.title)
			} else {
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
		}
	})

	t.Run("GetCurrentUser", func(t *testing.T) {
		fooUser := createRandomUser(t, lct.DB())

		tests := []struct {
			title              string
			reqUser            *model.User
			expectedStatusCode int
			expectedBody       message.ProfileResponse
			expectedError      map[string]interface{}
			hasError           bool
		}{
			{
				"get current user: success",
				fooUser,
				http.StatusOK,
				fooUser.ResponseProfile(false),
				nil,
				false,
			},
			{
				"get current user: wrong user id",
				&model.User{ID: 0},
				http.StatusNotFound,
				message.ProfileResponse{},
				map[string]interface{}{"error": "current user not found"},
				true,
			},
		}

		for _, tt := range tests {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
			w := httptest.NewRecorder()
			c, token := ctxWithToken(t, lct.Environ(), w, req, tt.reqUser.ID, time.Now().Add(-time.Hour))

			h.GetCurrentUser(c)

			actualCookies := w.Result().Header.Values("Set-Cookie")

			assert.Equal(t, tt.expectedStatusCode, w.Result().StatusCode, tt.title)

			if tt.hasError {
				actualBody := test.GetResponseBody[map[string]interface{}](t, w.Result())
				assert.Equal(t, tt.expectedError, actualBody, tt.title)
				assert.Empty(t, actualCookies, tt.title)
			} else {
				actualBody := test.GetResponseBody[message.ProfileResponse](t, w.Result())
				assert.Equal(t, tt.expectedBody, actualBody, tt.title)
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
		}
	})

	t.Run("UpdateCurrentUser", func(t *testing.T) {
		shortMaxLenString := strings.Repeat("a", 101)
		longMaxLenString := strings.Repeat("a", 256)
		passwordMaxLenString := strings.Repeat("a", 51)

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
			reqUser            *model.User
			reqBody            *message.UpdateUserRequest
			expectedStatusCode int
			expectedBody       message.ProfileResponse
			expectedError      map[string]interface{}
			hasError           bool
		}{
			{
				"update fooUser: success",
				fooUser,
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
				nil,
				false,
			},
			{
				"update fooUser: ignore zero-value field",
				fooUser,
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
				nil,
				false,
			},
			{
				"update: wrong user id",
				&model.User{ID: 0},
				&message.UpdateUserRequest{},
				http.StatusNotFound,
				message.ProfileResponse{},
				map[string]interface{}{"error": "current user not found"},
				true,
			},
			{
				"update: invalid username format",
				fooUser,
				&message.UpdateUserRequest{
					Username: "_invalid@@username_",
				},
				http.StatusBadRequest,
				message.ProfileResponse{},
				map[string]interface{}{"error": "validation error: Username: must be in a valid format."},
				true,
			},
			{
				"update: username is too short",
				fooUser,
				&message.UpdateUserRequest{
					Username: "abc",
				},
				http.StatusBadRequest,
				message.ProfileResponse{},
				map[string]interface{}{"error": "validation error: Username: the length must be between 5 and 100."},
				true,
			},
			{
				"update: username is too long",
				fooUser,
				&message.UpdateUserRequest{
					Username: shortMaxLenString,
				},
				http.StatusBadRequest,
				message.ProfileResponse{},
				map[string]interface{}{"error": "validation error: Username: the length must be between 5 and 100."},
				true,
			},
			{
				"update: password is too short",
				fooUser,
				&message.UpdateUserRequest{
					Password: "abc",
				},
				http.StatusBadRequest,
				message.ProfileResponse{},
				map[string]interface{}{"error": "validation error: Password: the length must be between 7 and 50."},
				true,
			},
			{
				"update: password is too long",
				fooUser,
				&message.UpdateUserRequest{
					Password: passwordMaxLenString,
				},
				http.StatusBadRequest,
				message.ProfileResponse{},
				map[string]interface{}{"error": "validation error: Password: the length must be between 7 and 50."},
				true,
			},
			{
				"update: weak password",
				fooUser,
				&message.UpdateUserRequest{
					Password: "password",
				},
				http.StatusBadRequest,
				message.ProfileResponse{},
				map[string]interface{}{"error": "validation error: Password: must have at least one uppercase, one number, one symbol or punctuation."},
				true,
			},
			{
				"update: email is too short",
				fooUser,
				&message.UpdateUserRequest{
					Email: "abc",
				},
				http.StatusBadRequest,
				message.ProfileResponse{},
				map[string]interface{}{"error": "validation error: Email: the length must be between 5 and 100."},
				true,
			},
			{
				"update: email is too long",
				fooUser,
				&message.UpdateUserRequest{
					Email: shortMaxLenString,
				},
				http.StatusBadRequest,
				message.ProfileResponse{},
				map[string]interface{}{"error": "validation error: Email: the length must be between 5 and 100."},
				true,
			},
			{
				"update: invalid email format",
				fooUser,
				&message.UpdateUserRequest{
					Email: "invalid_email_format",
				},
				http.StatusBadRequest,
				message.ProfileResponse{},
				map[string]interface{}{"error": "validation error: Email: must be a valid email address."},
				true,
			},
			{
				"update: name is too short",
				fooUser,
				&message.UpdateUserRequest{
					Name: "abc",
				},
				http.StatusBadRequest,
				message.ProfileResponse{},
				map[string]interface{}{"error": "validation error: Name: the length must be between 5 and 100."},
				true,
			},
			{
				"update: name is too long",
				fooUser,
				&message.UpdateUserRequest{
					Name: shortMaxLenString,
				},
				http.StatusBadRequest,
				message.ProfileResponse{},
				map[string]interface{}{"error": "validation error: Name: the length must be between 5 and 100."},
				true,
			},
			{
				"update: bio is too long",
				fooUser,
				&message.UpdateUserRequest{
					Bio: longMaxLenString,
				},
				http.StatusBadRequest,
				message.ProfileResponse{},
				map[string]interface{}{"error": "validation error: Bio: the length must be no more than 255."},
				true,
			},
			{
				"update: image is too long",
				fooUser,
				&message.UpdateUserRequest{
					Image: longMaxLenString,
				},
				http.StatusBadRequest,
				message.ProfileResponse{},
				map[string]interface{}{"error": "validation error: Image: the length must be no more than 255."},
				true,
			},
			{
				"update: invalid image url format",
				fooUser,
				&message.UpdateUserRequest{
					Image: "invalid_url_format",
				},
				http.StatusBadRequest,
				message.ProfileResponse{},
				map[string]interface{}{"error": "validation error: Image: must be a valid URL."},
				true,
			},
		}

		for _, tt := range tests {
			body, err := json.Marshal(tt.reqBody)
			if err != nil {
				t.Fatal(err)
			}

			req := httptest.NewRequest(http.MethodPut, "/api/v1/me", bytes.NewReader(body))
			w := httptest.NewRecorder()
			c, token := ctxWithToken(t, lct.Environ(), w, req, tt.reqUser.ID, time.Now().Add(-time.Hour))

			h.UpdateCurrentUser(c)

			actualCookies := w.Result().Header.Values("Set-Cookie")

			assert.Equal(t, tt.expectedStatusCode, w.Result().StatusCode, tt.title)

			if tt.hasError {
				actualBody := test.GetResponseBody[map[string]interface{}](t, w.Result())
				assert.Equal(t, tt.expectedError, actualBody, tt.title)
				assert.Empty(t, actualCookies, tt.title)
			} else {
				actualBody := test.GetResponseBody[message.ProfileResponse](t, w.Result())
				assert.Equal(t, tt.expectedBody, actualBody, tt.title)
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
		}
	})
}
