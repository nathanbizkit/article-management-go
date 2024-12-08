package middleware

import (
	"net/http"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nathanbizkit/article-management-go/auth"
	"github.com/nathanbizkit/article-management-go/test"
	"github.com/stretchr/testify/assert"
)

func TestUnit_AuthMiddleware(t *testing.T) {
	if !testing.Short() {
		t.Skip("skipping unit tests.")
	}

	gin.SetMode("test")

	environ := test.NewTestENV(t)
	l := test.NewTestLogger(t)
	authen := auth.New(environ)

	t.Run("Auth", func(t *testing.T) {
		token, err := authen.GenerateToken(10)
		if err != nil {
			t.Fatal(err)
		}

		tests := []struct {
			title              string
			strictCookie       bool
			reqHeaders         []header
			reqCookies         []*http.Cookie
			expectedStatusCode int
			expectedBody       map[string]interface{}
			expectedError      map[string]interface{}
			hasError           bool
		}{
			{
				"auth strict cookie: valid token",
				true,
				[]header{
					{
						Key:   "Origin",
						Value: "http://localhost:8000",
					},
				},
				[]*http.Cookie{
					{
						Name:     "session",
						Value:    url.QueryEscape(token.Token),
						MaxAge:   int((5 * (24 * time.Hour)).Seconds()),
						Path:     "/api/v1",
						Domain:   environ.AuthCookieDomain,
						SameSite: http.SameSiteStrictMode,
						Secure:   true,
						HttpOnly: true,
					},
				},
				http.StatusOK,
				map[string]interface{}{"user_id": "10"},
				nil,
				false,
			},
			{
				"auth with no strict cookie: allow public request",
				false,
				[]header{
					{
						Key:   "Origin",
						Value: "http://localhost:8000",
					},
				},
				nil,
				http.StatusOK,
				map[string]interface{}{"user_id": "0"},
				nil,
				false,
			},
			{
				"auth strict cookie: no session cookie",
				true,
				[]header{
					{
						Key:   "Origin",
						Value: "http://localhost:8000",
					},
				},
				nil,
				http.StatusUnauthorized,
				map[string]interface{}{"error": "unauthorized"},
				nil,
				false,
			},
		}

		for _, tt := range tests {
			router := gin.New()
			router.Use(Auth(&l, authen, tt.strictCookie))
			router.GET("/", func(ctx *gin.Context) {
				userID := authen.GetContextUserID(ctx)
				ctx.AbortWithStatusJSON(http.StatusOK, gin.H{"user_id": strconv.Itoa(int(userID))})
			})

			w := performRequest(t, router, http.MethodGet, "/", tt.reqHeaders, tt.reqCookies)

			actualBody := test.GetResponseBody[map[string]interface{}](t, w.Result())

			assert.Equal(t, tt.expectedStatusCode, w.Result().StatusCode, tt.title)

			if tt.hasError {
				assert.Equal(t, tt.expectedError, actualBody, tt.title)
			} else {
				assert.Equal(t, tt.expectedBody, actualBody, tt.title)
			}
		}
	})
}
