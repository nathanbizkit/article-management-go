package auth

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/nathanbizkit/article-management/env"
)

const (
	tokenTTL     = 72 * time.Hour
	refreshTTL   = 7 * (24 * time.Hour)
	cookieMaxAge = 5 * (24 * time.Hour)
)

type claims struct {
	UserID *uint `json:"user_id,omitempty"`
	jwt.RegisteredClaims
}

// AuthToken definition
type AuthToken struct {
	Token        string
	RefreshToken string
}

// Auth definition
type Auth struct {
	environ *env.ENV
}

// New returns a new auth with env
func New(environ *env.ENV) *Auth {
	return &Auth{environ: environ}
}

// GenerateToken generates a new auth token with expired date computed with current time
func (a *Auth) GenerateToken(id uint) (*AuthToken, error) {
	return a.GenerateTokenWithTime(id, time.Now())
}

// GenerateTokenWithTime generates a new auth token with expired date computed with specified time
func (a *Auth) GenerateTokenWithTime(id uint, t time.Time) (*AuthToken, error) {
	token, err := generateToken(a.environ.AuthJWTSecretKey, id, t, tokenTTL)
	if err != nil {
		return nil, err
	}

	refreshToken, err := generateToken(a.environ.AuthJWTSecretKey, id, t, refreshTTL)
	if err != nil {
		return nil, err
	}

	return &AuthToken{Token: token, RefreshToken: refreshToken}, nil
}

func generateToken(key string, id uint, now time.Time, d time.Duration) (string, error) {
	claims := &claims{
		&id,
		jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(d)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenString, err := token.SignedString([]byte(key))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// SetContextUserID sets auth user id to http context
func (a *Auth) SetContextUserID(ctx *gin.Context, id uint) {
	ctx.Set("auth_user_id", id)
}

// GetContextUserID returns auth user id from http context
func (a *Auth) GetContextUserID(ctx *gin.Context) uint {
	return ctx.GetUint("auth_user_id")
}

// GetUserID gets a user id from request context
func (a *Auth) GetUserID(ctx *gin.Context, strictCookie, refresh bool) (uint, error) {
	tokenName := "session"
	if refresh {
		tokenName = "refreshToken"
	}

	tokenString, err := ctx.Cookie(tokenName)
	if err != nil {
		if strictCookie {
			return 0, err
		}

		// allow unsecured connection
		return 0, nil
	}

	if tokenString == "" {
		if strictCookie {
			return 0, errors.New("auth token is empty")
		}

		// allow unsecured connection
		return 0, nil
	}

	token, err := jwt.ParseWithClaims(
		tokenString, &claims{},
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(a.environ.AuthJWTSecretKey), nil
		},
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
	)
	if err != nil {
		return 0, err
	}

	claims, ok := token.Claims.(*claims)
	if !ok || claims.UserID == nil {
		return 0, errors.New("invalid token claims")
	}

	return *claims.UserID, nil
}

// SetCookieToken sets a jwt token cookie in http header
func (a *Auth) SetCookieToken(ctx *gin.Context, token AuthToken, path string) {
	ctx.SetSameSite(http.SameSiteStrictMode)
	ctx.SetCookie(
		"session", token.Token,
		int(cookieMaxAge.Seconds()),
		path, a.environ.AuthCookieDomain, true, true,
	)
	ctx.SetCookie(
		"refreshToken", token.RefreshToken,
		int(cookieMaxAge.Seconds()),
		path, a.environ.AuthCookieDomain, true, true,
	)
}
