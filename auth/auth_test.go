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

	environ := test.NewTestENV(t)
	authen := New(environ)

	t.Run("GenerateToken", func(t *testing.T) {
		id := uint(10)
		actual, err := authen.GenerateToken(id)

		assert.NoError(t, err)
		assert.NotEmpty(t, actual)
		assert.NotEmpty(t, actual.Token)
		assert.NotEmpty(t, actual.RefreshToken)
		assert.Equal(t, id, parseToken(t, actual.Token, environ.AuthJWTSecretKey))
		assert.Equal(t, id, parseToken(t, actual.RefreshToken, environ.AuthJWTSecretKey))
	})

	t.Run("ContextUserID: Set & Get", func(t *testing.T) {
		id := uint(10)
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		authen.SetContextUserID(ctx, id)
		assert.Equal(t, id, authen.GetContextUserID(ctx))
	})

	t.Run("GetUserID", func(t *testing.T) {
		id := uint(10)
		token, err := authen.GenerateToken(id)
		if err != nil {
			t.Fatal(err)
		}

		zeroTime := time.Date(0, 0, 0, 0, 0, 0, 0, time.Local)
		expiredToken, err := authen.GenerateTokenWithTime(id, zeroTime)
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
			SignedString([]byte(environ.AuthJWTSecretKey))
		if err != nil {
			t.Fatal(err)
		}

		normalCtx := func() *gin.Context {
			ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
			ctx.Request = &http.Request{
				Header: make(http.Header),
			}
			test.AddCookieToRequest(
				t, ctx.Request,
				"session", token.Token, environ.AuthCookieDomain,
			)
			test.AddCookieToRequest(
				t, ctx.Request,
				"refreshToken", token.RefreshToken, environ.AuthCookieDomain,
			)
			return ctx
		}

		noCookieCtx := func() *gin.Context {
			ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
			ctx.Request = &http.Request{
				Header: make(http.Header),
			}
			return ctx
		}

		emptyTokenCtx := func() *gin.Context {
			ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
			ctx.Request = &http.Request{
				Header: make(http.Header),
			}
			test.AddCookieToRequest(t, ctx.Request, "session", "", environ.AuthCookieDomain)
			test.AddCookieToRequest(t, ctx.Request, "refreshToken", "", environ.AuthCookieDomain)
			return ctx
		}

		expiredTokenCtx := func() *gin.Context {
			ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
			ctx.Request = &http.Request{
				Header: make(http.Header),
			}
			test.AddCookieToRequest(
				t, ctx.Request,
				"session", expiredToken.Token, environ.AuthCookieDomain,
			)
			test.AddCookieToRequest(
				t, ctx.Request, "refreshToken",
				expiredToken.RefreshToken, environ.AuthCookieDomain,
			)
			return ctx
		}

		rsaTokenCtx := func() *gin.Context {
			ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
			ctx.Request = &http.Request{
				Header: make(http.Header),
			}
			test.AddCookieToRequest(
				t, ctx.Request,
				"session", rsaToken, environ.AuthCookieDomain,
			)
			test.AddCookieToRequest(
				t, ctx.Request,
				"refreshToken", rsaToken, environ.AuthCookieDomain,
			)
			return ctx
		}

		emptyClaimTokenCtx := func() *gin.Context {
			ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
			ctx.Request = &http.Request{
				Header: make(http.Header),
			}
			test.AddCookieToRequest(
				t, ctx.Request,
				"session", emptyClaimToken, environ.AuthCookieDomain,
			)
			test.AddCookieToRequest(
				t, ctx.Request,
				"refreshToken", emptyClaimToken, environ.AuthCookieDomain,
			)
			return ctx
		}

		tests := []struct {
			title          string
			getContextFn   func() *gin.Context
			strictCookie   bool
			refresh        bool
			expectedUserID uint
			hasError       bool
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
				"get user id: allow public connection if no cookie is found",
				noCookieCtx,
				false,
				false,
				0,
				false,
			},
			{
				"get user id: allow public connection if cookie is empty",
				emptyTokenCtx,
				false,
				false,
				0,
				false,
			},
		}

		for _, tt := range tests {
			actualUserID, err := authen.GetUserID(tt.getContextFn(), tt.strictCookie, tt.refresh)

			if tt.hasError {
				assert.Error(t, err, tt.title)
			} else {
				assert.NoError(t, err, tt.title)
			}

			assert.Equal(t, tt.expectedUserID, actualUserID, tt.title)
		}
	})

	t.Run("SetCookieToken", func(t *testing.T) {
		id := uint(10)
		token, err := authen.GenerateToken(id)
		if err != nil {
			t.Fatal(err)
		}

		w1 := httptest.NewRecorder()
		test.AddCookieToResponse(t, w1, "session", token.Token, environ.AuthCookieDomain)
		test.AddCookieToResponse(t, w1, "refreshToken", token.RefreshToken, environ.AuthCookieDomain)
		expected := w1.Header().Get("Set-Cookie")

		w2 := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w2)

		authen.SetCookieToken(ctx, *token, "/api/v1")
		actual := w2.Header().Get("Set-Cookie")

		assert.Equal(t, expected, actual)
	})
}

func parseToken(t *testing.T, tokenString, secretKey string) uint {
	t.Helper()

	token, err := jwt.ParseWithClaims(
		tokenString, &claims{},
		func(t *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		},
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
	)
	if err != nil {
		t.Fatalf("failed to parse jwt token: %s", err)
	}

	if !token.Valid {
		t.Fatalf("invalid token: %s", err)
	}

	claims, ok := token.Claims.(*claims)
	if !ok {
		t.Fatal("cannot map token to claims")
	}

	return *claims.UserID
}
