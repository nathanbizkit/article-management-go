package test

import (
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/nathanbizkit/article-management/env"
	"github.com/nathanbizkit/article-management/test/container"
	"github.com/rs/zerolog"
)

const englishCharset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var ltc *container.LocalTestContainer
var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

// NewTestENV returns an env for testing
func NewTestENV(t *testing.T) *env.ENV {
	t.Helper()

	return &env.ENV{
		AppMode:          "test",
		AppPort:          "8000",
		AuthJWTSecretKey: "secretkey",
		AuthCookieDomain: "localhost",
		DBUser:           "root",
		DBPass:           "password",
		DBHost:           "db_test",
		DBPort:           "5432",
		DBName:           "app_test",
	}
}

// NewTestLogger returns a logger for testing
func NewTestLogger(t *testing.T) zerolog.Logger {
	t.Helper()

	w := zerolog.ConsoleWriter{Out: io.Discard}
	// w := zerolog.ConsoleWriter{Out: os.Stderr}
	return zerolog.New(w).With().Timestamp().Caller().Logger()
}

// NewLocalTestContainer returns a local test container
func NewLocalTestContainer(t *testing.T) *container.LocalTestContainer {
	t.Helper()

	if testing.Short() {
		return nil
	}

	ltc, err := container.NewLocalTestContainer()
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		ltc.Close()
	})

	return ltc
}

// RandomString returns a random string with x length in English and Numbers
func RandomString(t *testing.T, length int) string {
	t.Helper()

	b := make([]byte, length)
	for i := range b {
		b[i] = englishCharset[seededRand.Intn(len(englishCharset))]
	}

	return string(b)
}

// AddCookieToRequest attaches a api-related cookie to request header
func AddCookieToRequest(t *testing.T, req *http.Request, name, value, domain string) {
	t.Helper()

	cookie := &http.Cookie{
		Name:     name,
		Value:    url.QueryEscape(value),
		MaxAge:   int((20 * (24 * time.Hour)).Seconds()),
		Path:     "/api/v1",
		Domain:   domain,
		SameSite: http.SameSiteStrictMode,
		Secure:   true,
		HttpOnly: true,
	}

	if v := cookie.String(); v != "" {
		req.Header.Add("Cookie", v)
	}
}

// AddCookieToResponse attaches a api-related cookie to response header
func AddCookieToResponse(t *testing.T, w http.ResponseWriter, name, value, domain string) {
	t.Helper()

	cookie := &http.Cookie{
		Name:     name,
		Value:    url.QueryEscape(value),
		MaxAge:   int((20 * (24 * time.Hour)).Seconds()),
		Path:     "/api/v1",
		Domain:   domain,
		SameSite: http.SameSiteStrictMode,
		Secure:   true,
		HttpOnly: true,
	}

	if v := cookie.String(); v != "" {
		w.Header().Add("Set-Cookie", v)
	}
}
