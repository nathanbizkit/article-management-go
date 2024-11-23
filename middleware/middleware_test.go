package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type header struct {
	Key   string
	Value string
}

func performRequest(t *testing.T, r http.Handler, method, path string, headers []header, cookies []*http.Cookie) *httptest.ResponseRecorder {
	t.Helper()

	if path == "" {
		path = "/"
	}

	req := httptest.NewRequest(method, path, nil)

	for _, header := range headers {
		req.Header.Add(header.Key, header.Value)
	}

	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	return w
}
