package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nathanbizkit/article-management/env"
	"github.com/nathanbizkit/article-management/test"
	"github.com/stretchr/testify/assert"
)

func TestUnit_SecureMiddleware(t *testing.T) {
	if !testing.Short() {
		t.Skip("skipping unit tests.")
	}

	gin.SetMode("test")

	t.Run("Secure", func(t *testing.T) {
		tests := []struct {
			title    string
			getEnvFn func() *env.ENV
			setupFn  func(r *gin.Engine)
			runFn    func(r *gin.Engine) *httptest.ResponseRecorder
			testFn   func(t *testing.T, w *httptest.ResponseRecorder, title string)
		}{
			{
				"secure (development): no redirect",
				func() *env.ENV {
					environ := test.NewTestENV(t)
					environ.AppTLSPort = "8443"
					environ.IsDevelopment = true
					return environ
				},
				func(r *gin.Engine) {
					r.GET("/", func(ctx *gin.Context) {
						ctx.AbortWithStatus(http.StatusNoContent)
					})
				},
				func(r *gin.Engine) *httptest.ResponseRecorder {
					headers := []header{
						{
							Key:   "Origin",
							Value: "http://localhost:8000",
						},
					}
					return performRequest(t, r, http.MethodGet, "/", headers, nil)
				},
				func(t *testing.T, w *httptest.ResponseRecorder, title string) {
					assert.Equal(t, http.StatusNoContent, w.Result().StatusCode, title)
				},
			},
			{
				"secure (production): redirect http to https",
				func() *env.ENV {
					environ := test.NewTestENV(t)
					environ.AppTLSPort = "8443"
					environ.IsDevelopment = false
					return environ
				},
				func(r *gin.Engine) {
					r.GET("/", func(ctx *gin.Context) {
						ctx.AbortWithStatus(http.StatusNoContent)
					})
				},
				func(r *gin.Engine) *httptest.ResponseRecorder {
					headers := []header{
						{
							Key:   "Origin",
							Value: "http://localhost:8000",
						},
					}
					return performRequest(t, r, http.MethodGet, "/", headers, nil)
				},
				func(t *testing.T, w *httptest.ResponseRecorder, title string) {
					assert.Equal(t, http.StatusMovedPermanently, w.Result().StatusCode, title)
					assert.Equal(t, w.Result().Header.Get("Location"), "https://example.com/", title)
				},
			},
		}

		for _, tt := range tests {
			router := gin.New()
			router.Use(Secure(tt.getEnvFn()))

			tt.setupFn(router)
			tt.testFn(t, tt.runFn(router), tt.title)
		}
	})
}
