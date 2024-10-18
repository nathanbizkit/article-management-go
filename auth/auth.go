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
	tokenTTLInHour     = 72 * time.Hour
	refreshTTLInHour   = 14 * (24 * time.Hour)
	cookieMaxAgeInHour = (20 * (25 * time.Hour))
)

type claims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

// AuthToken definition
type AuthToken struct {
	Token        string
	RefreshToken string
}

// Auth is an authentication service
type Auth struct {
	env *env.ENV
}

// New returns a new auth with env
func New(e *env.ENV) *Auth {
	return &Auth{env: e}
}

// GenerateToken generates a new auth token
func (a *Auth) GenerateToken(id uint) (*AuthToken, error) {
	token, err := generateToken(a.env.AuthJWTSecretKey, id, time.Now(), tokenTTLInHour)
	if err != nil {
		return nil, err
	}

	rt, err := generateToken(a.env.AuthJWTSecretKey, id, time.Now(), refreshTTLInHour)
	if err != nil {
		return nil, err
	}

	return &AuthToken{Token: token, RefreshToken: rt}, nil
}

// GenerateTokenWithTime generates a new auth token with expired date computed with specified time
func (a *Auth) GenerateTokenWithTime(id uint, t time.Time) (*AuthToken, error) {
	// for testing purposes
	token, err := generateToken(a.env.AuthJWTSecretKey, id, t, tokenTTLInHour)
	if err != nil {
		return nil, err
	}

	rt, err := generateToken(a.env.AuthJWTSecretKey, id, t, refreshTTLInHour)
	if err != nil {
		return nil, err
	}

	return &AuthToken{Token: token, RefreshToken: rt}, nil
}

func generateToken(key string, id uint, now time.Time, fromNow time.Duration) (string, error) {
	claims := &claims{
		id,
		jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(fromNow)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	t, err := token.SignedString([]byte(key))
	if err != nil {
		return "", err
	}

	return t, nil
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
func (a *Auth) GetUserID(ctx *gin.Context, refresh bool) (uint, error) {
	tokenName := "session"
	if refresh {
		tokenName = "refreshToken"
	}

	tokenString, err := ctx.Cookie(tokenName)
	if err != nil {
		return 0, err
	}

	token, err := jwt.ParseWithClaims(
		tokenString, &claims{},
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(a.env.AuthJWTSecretKey), nil
		},
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
	)
	if err != nil {
		return 0, err
	}

	if !token.Valid {
		return 0, errors.New("invalid token")
	}

	c, ok := token.Claims.(*claims)
	if !ok {
		return 0, errors.New("invalid: cannot map token to claims")
	}

	return c.UserID, nil
}

// SetCookieToken sets a jwt token cookie in http header
func (a *Auth) SetCookieToken(ctx *gin.Context, t AuthToken, path string) {
	ctx.SetSameSite(http.SameSiteStrictMode)
	ctx.SetCookie("session", t.Token, int(cookieMaxAgeInHour.Seconds()),
		path, a.env.AuthCookieDomain, true, true)
	ctx.SetCookie("refreshToken", t.RefreshToken, int(cookieMaxAgeInHour.Seconds()),
		path, a.env.AuthCookieDomain, true, true)
}

// GetCookieToken returns a jwt token in http cookie
func (a *Auth) GetCookieToken(ctx *gin.Context) (*AuthToken, error) {
	t, err := ctx.Cookie("session")
	if err != nil {
		return nil, err
	}

	rt, err := ctx.Cookie("refreshToken")
	if err != nil {
		return nil, err
	}

	return &AuthToken{Token: t, RefreshToken: rt}, nil
}
