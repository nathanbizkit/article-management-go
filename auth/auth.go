package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/nathanbizkit/article-management/env"
)

type claims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

type Auth struct {
	e *env.ENV
}

// New returns a new auth with env
func New(e *env.ENV) *Auth {
	return &Auth{e: e}
}

// GenerateToken generates a new token
func (a *Auth) GenerateToken(id uint) (string, error) {
	return generateToken(a.e.AuthJWTSecretKey, id, time.Now(), 72*time.Hour)
}

// GenerateTokenWithTime generates a new token with expired date computed with specified time
func (a *Auth) GenerateTokenWithTime(id uint, t time.Time) (string, error) {
	// for testing purposes
	return generateToken(a.e.AuthJWTSecretKey, id, t, 72*time.Hour)
}

// GenerateRefreshToken generates a new refresh token
func (a *Auth) GenerateRefreshToken(id uint) (string, error) {
	return generateToken(a.e.AuthJWTSecretKey, id, time.Now(), 72*(24*time.Hour))
}

// GenerateRefreshTokenWithTime generates a new refresh token with expired date computed with specified time
func (a *Auth) GenerateRefreshTokenWithTime(id uint, t time.Time) (string, error) {
	// for testing purposes
	return generateToken(a.e.AuthJWTSecretKey, id, t, 72*(24*time.Hour))
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

// GetUserID gets a user id from a token
func (a *Auth) GetUserID(token string) (uint, error) {
	parsed, err := jwt.ParseWithClaims(
		token, &claims{},
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(a.e.AuthJWTSecretKey), nil
		},
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
	)
	if err != nil {
		return 0, err
	}

	if !parsed.Valid {
		return 0, errors.New("invalid token")
	}

	c, ok := parsed.Claims.(*claims)
	if !ok {
		return 0, errors.New("invalid: cannot map token to claims")
	}

	return c.UserID, nil
}
