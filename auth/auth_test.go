package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/nathanbizkit/article-management/test"
	"github.com/stretchr/testify/assert"
)

func TestUnit_Auth(t *testing.T) {
	if !testing.Short() {
		t.Skip("skipping unit tests.")
	}

	gin.SetMode("test")

	e := test.NewTestENV(t)
	a := New(e)

	t.Run("GenerateToken", func(t *testing.T) {
		id := uint(10)
		actual, err := a.GenerateToken(id)

		assert.NoError(t, err)
		assert.NotEmpty(t, actual)
		assert.NotEmpty(t, actual.Token)
		assert.NotEmpty(t, actual.RefreshToken)
		assert.Equal(t, id, parseToken(t, actual.Token, e.AuthJWTSecretKey))
		assert.Equal(t, id, parseToken(t, actual.RefreshToken, e.AuthJWTSecretKey))
	})

	t.Run("ContextUserID: Set & Get", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		a.SetContextUserID(c, 10)
		assert.Equal(t, uint(10), a.GetContextUserID(c))
	})

	t.Run("GetUserID", func(t *testing.T) {
		id := uint(10)
		token, err := a.GenerateToken(id)
		if err != nil {
			t.Fatal(err)
		}

		zeroTime := time.Date(0, 0, 0, 0, 0, 0, 0, time.Local)
		expiredToken, err := a.GenerateTokenWithTime(id, zeroTime)
		if err != nil {
			t.Fatal(err)
		}

		privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
		if err != nil {
			t.Fatal(err)
		}

		rsaToken, err := jwt.NewWithClaims(jwt.SigningMethodRS512, &jwt.RegisteredClaims{}).
			SignedString(privateKey)
		if err != nil {
			t.Fatal(err)
		}

		emptyClaimToken, err := jwt.NewWithClaims(
			jwt.SigningMethodHS512,
			&jwt.RegisteredClaims{
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			}).
			SignedString([]byte(e.AuthJWTSecretKey))
		if err != nil {
			t.Fatal(err)
		}

		normalCtx := func() *gin.Context {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = &http.Request{
				Header: make(http.Header),
			}
			test.AddCookieToRequest(
				t, c.Request,
				"session", token.Token, e.AuthCookieDomain,
			)
			test.AddCookieToRequest(
				t, c.Request,
				"refreshToken", token.RefreshToken, e.AuthCookieDomain,
			)
			return c
		}

		noCookieCtx := func() *gin.Context {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = &http.Request{
				Header: make(http.Header),
			}
			return c
		}

		emptyTokenCtx := func() *gin.Context {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = &http.Request{
				Header: make(http.Header),
			}
			test.AddCookieToRequest(t, c.Request, "session", "", e.AuthCookieDomain)
			test.AddCookieToRequest(t, c.Request, "refreshToken", "", e.AuthCookieDomain)
			return c
		}

		expiredTokenCtx := func() *gin.Context {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = &http.Request{
				Header: make(http.Header),
			}
			test.AddCookieToRequest(
				t, c.Request,
				"session", expiredToken.Token, e.AuthCookieDomain,
			)
			test.AddCookieToRequest(
				t, c.Request, "refreshToken",
				expiredToken.RefreshToken, e.AuthCookieDomain,
			)
			return c
		}

		rsaTokenCtx := func() *gin.Context {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = &http.Request{
				Header: make(http.Header),
			}
			test.AddCookieToRequest(
				t, c.Request,
				"session", rsaToken, e.AuthCookieDomain,
			)
			test.AddCookieToRequest(
				t, c.Request,
				"refreshToken", rsaToken, e.AuthCookieDomain,
			)
			return c
		}

		emptyClaimTokenCtx := func() *gin.Context {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = &http.Request{
				Header: make(http.Header),
			}
			test.AddCookieToRequest(
				t, c.Request,
				"session", emptyClaimToken, e.AuthCookieDomain,
			)
			test.AddCookieToRequest(
				t, c.Request,
				"refreshToken", emptyClaimToken, e.AuthCookieDomain,
			)
			return c
		}

		tests := []struct {
			title        string
			ctx          func() *gin.Context
			strictCookie bool
			refresh      bool
			expected     uint
			hasError     bool
		}{
			{
				"get user id (session): success",
				normalCtx,
				true,
				false,
				id,
				false,
			},
			{
				"get user id (refresh): success",
				normalCtx,
				true,
				true,
				id,
				false,
			},
			{
				"get user id (session): no token in cookie",
				noCookieCtx,
				true,
				false,
				0,
				true,
			},
			{
				"get user id (refresh): no token in cookie",
				noCookieCtx,
				true,
				true,
				0,
				true,
			},
			{
				"get user id (session): empty token",
				emptyTokenCtx,
				true,
				false,
				0,
				true,
			},
			{
				"get user id (refresh): empty token",
				emptyTokenCtx,
				true,
				true,
				0,
				true,
			},
			{
				"get user id (session): expired token",
				expiredTokenCtx,
				true,
				false,
				0,
				true,
			},
			{
				"get user id (refresh): expired token",
				expiredTokenCtx,
				true,
				true,
				0,
				true,
			},
			{
				"get user id (session): wrong jwt signing method",
				rsaTokenCtx,
				true,
				false,
				0,
				true,
			},
			{
				"get user id (refresh): wrong jwt signing method",
				rsaTokenCtx,
				true,
				true,
				0,
				true,
			},
			{
				"get user id (session): empty token claims",
				emptyClaimTokenCtx,
				true,
				false,
				0,
				true,
			},
			{
				"get user id (refresh): empty token claims",
				emptyClaimTokenCtx,
				true,
				true,
				0,
				true,
			},
			{
				"get user id: allow unsecured connection",
				noCookieCtx,
				false,
				false,
				0,
				false,
			},
		}

		for _, tt := range tests {
			actual, err := a.GetUserID(tt.ctx(), tt.strictCookie, tt.refresh)

			if tt.hasError {
				assert.Error(t, err, tt.title)
			} else {
				assert.NoError(t, err, tt.title)
			}

			assert.Equal(t, tt.expected, actual, tt.title)
		}
	})

	t.Run("SetCookieToken", func(t *testing.T) {
		token, err := a.GenerateToken(10)
		if err != nil {
			t.Fatal(err)
		}

		w1 := httptest.NewRecorder()
		test.AddCookieToResponse(t, w1, "session", token.Token, e.AuthCookieDomain)
		test.AddCookieToResponse(t, w1, "refreshToken", token.RefreshToken, e.AuthCookieDomain)
		expected := w1.Header().Get("Set-Cookie")

		w2 := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w2)

		a.SetCookieToken(c, *token, "/api/v1")

		actual := w2.Header().Get("Set-Cookie")
		assert.Equal(t, expected, actual)
	})

	t.Run("GetCookieToken", func(t *testing.T) {
		id := uint(10)
		token, err := a.GenerateToken(id)
		if err != nil {
			t.Fatal(err)
		}

		tests := []struct {
			title    string
			ctx      func() *gin.Context
			expected *AuthToken
			hasError bool
		}{
			{
				"get cookie token: success",
				func() *gin.Context {
					c, _ := gin.CreateTestContext(httptest.NewRecorder())
					c.Request = &http.Request{
						Header: make(http.Header),
					}
					test.AddCookieToRequest(
						t, c.Request,
						"session", token.Token, e.AuthCookieDomain,
					)
					test.AddCookieToRequest(
						t, c.Request,
						"refreshToken", token.RefreshToken, e.AuthCookieDomain,
					)
					return c
				},
				token,
				false,
			},
			{
				"get cookie token (session): no token in cookie",
				func() *gin.Context {
					c, _ := gin.CreateTestContext(httptest.NewRecorder())
					c.Request = &http.Request{
						Header: make(http.Header),
					}
					return c
				},
				nil,
				true,
			},
			{
				"get cookie token (refresh): no token in cookie",
				func() *gin.Context {
					c, _ := gin.CreateTestContext(httptest.NewRecorder())
					c.Request = &http.Request{
						Header: make(http.Header),
					}
					test.AddCookieToRequest(
						t, c.Request,
						"session", token.Token, e.AuthCookieDomain,
					)
					return c
				},
				nil,
				true,
			},
			{
				"get cookie token (session): empty token",
				func() *gin.Context {
					c, _ := gin.CreateTestContext(httptest.NewRecorder())
					c.Request = &http.Request{
						Header: make(http.Header),
					}
					test.AddCookieToRequest(
						t, c.Request,
						"session", "", e.AuthCookieDomain,
					)
					test.AddCookieToRequest(
						t, c.Request,
						"refreshToken", token.RefreshToken, e.AuthCookieDomain,
					)
					return c
				},
				nil,
				true,
			},
			{
				"get cookie token (refresh): empty token",
				func() *gin.Context {
					c, _ := gin.CreateTestContext(httptest.NewRecorder())
					c.Request = &http.Request{
						Header: make(http.Header),
					}
					test.AddCookieToRequest(
						t, c.Request,
						"session", token.Token, e.AuthCookieDomain,
					)
					test.AddCookieToRequest(
						t, c.Request,
						"refreshToken", "", e.AuthCookieDomain,
					)
					return c
				},
				nil,
				true,
			},
		}

		for _, tt := range tests {
			actual, err := a.GetCookieToken(tt.ctx())

			if tt.hasError {
				assert.Error(t, err, tt.title)
			} else {
				assert.NoError(t, err, tt.title)
			}

			assert.Equal(t, tt.expected, actual, tt.title)
		}
	})
}

func parseToken(t *testing.T, token, secretKey string) uint {
	t.Helper()

	tk, err := jwt.ParseWithClaims(
		token, &claims{},
		func(t *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		},
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
	)
	if err != nil {
		t.Fatalf("failed to parse jwt token: %s", err)
	}

	if !tk.Valid {
		t.Fatalf("invalid token: %s", err)
	}

	c, ok := tk.Claims.(*claims)
	if !ok {
		t.Fatal("cannot map token to claims")
	}

	return *c.UserID
}
