package test

import (
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/nathanbizkit/article-management-go/env"
	"github.com/nathanbizkit/article-management-go/test/container"
	"github.com/rs/zerolog"
)

const englishCharset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

// NewTestENV returns a mock of env object
func NewTestENV(t *testing.T) *env.ENV {
	t.Helper()

	return &env.ENV{
		AppMode:          "test",
		AppPort:          "8000",
		AppTLSPort:       "8443",
		AuthJWTSecretKey: "secretkey",
		AuthCookieDomain: "localhost",
		DBUser:           "root",
		DBPass:           "password",
		DBHost:           "db_test",
		DBPort:           "5432",
		DBName:           "app_test",
		IsDevelopment:    true,
	}
}

// NewTestLogger returns a logger for testing
func NewTestLogger(t *testing.T) zerolog.Logger {
	t.Helper()

	w := zerolog.ConsoleWriter{Out: io.Discard}
	return zerolog.New(w).With().Timestamp().Caller().Logger()
}

// NewLocalTestContainer returns a local test container
func NewLocalTestContainer(t *testing.T) *container.LocalTestContainer {
	t.Helper()

	if testing.Short() {
		t.Fatal("only available in integration tests")
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

	b := make([]byte, 0, length)
	for i := 0; i < length; i++ {
		b = append(b, englishCharset[seededRand.Intn(len(englishCharset))])
	}

	return string(b)
}

// AddCookieToRequest attaches a api-related cookie to request header
func AddCookieToRequest(t *testing.T, req *http.Request, name, value, domain string) {
	t.Helper()

	cookie := &http.Cookie{
		Name:     name,
		Value:    url.QueryEscape(value),
		MaxAge:   int((7 * (24 * time.Hour)).Seconds()),
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
		MaxAge:   int((7 * (24 * time.Hour)).Seconds()),
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

// GetResponseBody parses http json response body into specific object type
func GetResponseBody[T any](t *testing.T, res *http.Response) T {
	t.Helper()

	var jsonBody T
	err := json.NewDecoder(res.Body).Decode(&jsonBody)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	return jsonBody
}
