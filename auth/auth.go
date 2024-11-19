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
	cookieMaxAgeInHour = (20 * (24 * time.Hour))

	contextAuthUserID = "auth_user_id"
	cookieAuthSession = "session"
	cookieAuthRefresh = "refreshToken"
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
		&id,
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
	ctx.Set(contextAuthUserID, id)
}

// GetContextUserID returns auth user id from http context
func (a *Auth) GetContextUserID(ctx *gin.Context) uint {
	return ctx.GetUint(contextAuthUserID)
}

// GetUserID gets a user id from request context
func (a *Auth) GetUserID(ctx *gin.Context, secure, refresh bool) (uint, error) {
	tokenName := cookieAuthSession
	if refresh {
		tokenName = cookieAuthRefresh
	}

	tokenString, err := ctx.Cookie(tokenName)
	if err != nil {
		if secure {
			return 0, err
		}

		// allow unsecured connection if secure=false
		return 0, nil
	}

	if tokenString == "" {
		if secure {
			return 0, errors.New("auth token is empty")
		}

		// allow unsecured connection if secure=false
		return 0, nil
	}

	if tokenString == "" && secure {
		return 0, errors.New("auth token is empty")
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

	c, ok := token.Claims.(*claims)
	if !ok || c.UserID == nil {
		return 0, errors.New("invalid token claims")
	}

	return *c.UserID, nil
}

// SetCookieToken sets a jwt token cookie in http header
func (a *Auth) SetCookieToken(ctx *gin.Context, t AuthToken, path string) {
	ctx.SetSameSite(http.SameSiteStrictMode)
	ctx.SetCookie(
		cookieAuthSession, t.Token,
		int(cookieMaxAgeInHour.Seconds()),
		path, a.env.AuthCookieDomain, true, true,
	)
	ctx.SetCookie(
		cookieAuthRefresh, t.RefreshToken,
		int(cookieMaxAgeInHour.Seconds()),
		path, a.env.AuthCookieDomain, true, true,
	)
}

// GetCookieToken returns a jwt token in http cookie
func (a *Auth) GetCookieToken(ctx *gin.Context) (*AuthToken, error) {
	t, err := ctx.Cookie(cookieAuthSession)
	if err != nil {
		return nil, err
	}

	if t == "" {
		return nil, fmt.Errorf("%s is empty", cookieAuthSession)
	}

	rt, err := ctx.Cookie(cookieAuthRefresh)
	if err != nil {
		return nil, err
	}

	if rt == "" {
		return nil, fmt.Errorf("%s is empty", cookieAuthRefresh)
	}

	return &AuthToken{Token: t, RefreshToken: rt}, nil
}
