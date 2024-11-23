package middleware

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nathanbizkit/article-management/env"
	"github.com/nathanbizkit/article-management/test"
	"github.com/stretchr/testify/assert"
)

func TestUnit_CORSMiddleware(t *testing.T) {
	if !testing.Short() {
		t.Skip("skipping unit tests.")
	}

	gin.SetMode("test")

	t.Run("CORS", func(t *testing.T) {
		tests := []struct {
			title    string
			getEnvFn func() *env.ENV
			setupFn  func(r *gin.Engine)
			runFn    func(r *gin.Engine) *httptest.ResponseRecorder
			testFn   func(t *testing.T, w *httptest.ResponseRecorder, title string)
		}{
			{
				"cors (OPTIONS): allow all origins",
				func() *env.ENV {
					environ := test.NewTestENV(t)
					environ.CORSAllowedOrigins = []string{}
					return environ
				},
				func(r *gin.Engine) {},
				func(r *gin.Engine) *httptest.ResponseRecorder {
					headers := []header{
						{
							Key:   "Origin",
							Value: "http://localhost:8000",
						},
					}
					return performRequest(t, r, http.MethodOptions, "/", headers, nil)
				},
				func(t *testing.T, w *httptest.ResponseRecorder, title string) {
					assert.Equal(t, http.StatusNoContent, w.Result().StatusCode, title)
					assert.Equal(t,
						"true",
						w.Result().Header.Get("Access-Control-Allow-Credentials"),
						title,
					)
					assert.Equal(t,
						"*",
						w.Result().Header.Get("Access-Control-Allow-Origin"),
						title,
					)
					assert.Equal(t,
						"GET,POST,PUT,PATCH,DELETE,HEAD,OPTIONS",
						w.Result().Header.Get("Access-Control-Allow-Methods"),
						title,
					)
					assert.Equal(t,
						"Origin,Content-Length,Content-Type",
						w.Result().Header.Get("Access-Control-Allow-Headers"),
						title,
					)
					assert.Equal(t,
						strconv.FormatInt(int64((12*time.Hour)/time.Second), 10),
						w.Header().Get("Access-Control-Max-Age"),
						title,
					)
				},
			},
			{
				"cors (OPTIONS): allow fixed origins",
				func() *env.ENV {
					environ := test.NewTestENV(t)
					environ.CORSAllowedOrigins = []string{"http://localhost:8000"}
					return environ
				},
				func(r *gin.Engine) {},
				func(r *gin.Engine) *httptest.ResponseRecorder {
					headers := []header{
						{
							Key:   "Origin",
							Value: "http://localhost:8000",
						},
					}
					return performRequest(t, r, http.MethodOptions, "/", headers, nil)
				},
				func(t *testing.T, w *httptest.ResponseRecorder, title string) {
					assert.Equal(t, http.StatusNoContent, w.Result().StatusCode, title)
					assert.Equal(t,
						"true",
						w.Result().Header.Get("Access-Control-Allow-Credentials"),
						title,
					)
					assert.Equal(t,
						"http://localhost:8000",
						w.Result().Header.Get("Access-Control-Allow-Origin"),
						title,
					)
					assert.Equal(t,
						"GET,POST,PUT,PATCH,DELETE,HEAD,OPTIONS",
						w.Result().Header.Get("Access-Control-Allow-Methods"),
						title,
					)
					assert.Equal(t,
						"Origin,Content-Length,Content-Type",
						w.Result().Header.Get("Access-Control-Allow-Headers"),
						title,
					)
					assert.Equal(t,
						strconv.FormatInt(int64((12*time.Hour)/time.Second), 10),
						w.Header().Get("Access-Control-Max-Age"),
						title,
					)
				},
			},
			{
				"cors (OPTIONS): wrong origins ",
				func() *env.ENV {
					environ := test.NewTestENV(t)
					environ.CORSAllowedOrigins = []string{"http://localhost:8000"}
					return environ
				},
				func(r *gin.Engine) {},
				func(r *gin.Engine) *httptest.ResponseRecorder {
					headers := []header{
						{
							Key:   "Origin",
							Value: "http://localhost:3000",
						},
					}
					return performRequest(t, r, http.MethodOptions, "/", headers, nil)
				},
				func(t *testing.T, w *httptest.ResponseRecorder, title string) {
					assert.Equal(t, http.StatusForbidden, w.Result().StatusCode, title)
					assert.Empty(t, w.Header().Get("Access-Control-Allow-Credentials"), title)
					assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"), title)
					assert.Empty(t, w.Header().Get("Access-Control-Allow-Methods"), title)
					assert.Empty(t, w.Header().Get("Access-Control-Allow-Headers"), title)
					assert.Empty(t, w.Header().Get("Access-Control-Max-Age"), title)
				},
			},
			{
				"cors (GET): allow all origins",
				func() *env.ENV {
					environ := test.NewTestENV(t)
					environ.CORSAllowedOrigins = []string{}
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
					assert.Equal(t,
						"true",
						w.Result().Header.Get("Access-Control-Allow-Credentials"),
						title,
					)
					assert.Equal(t,
						"*",
						w.Result().Header.Get("Access-Control-Allow-Origin"),
						title,
					)
				},
			},
			{
				"cors (GET): wrong origins ",
				func() *env.ENV {
					environ := test.NewTestENV(t)
					environ.CORSAllowedOrigins = []string{"http://localhost:8000"}
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
							Value: "http://localhost:3000",
						},
					}
					return performRequest(t, r, http.MethodGet, "/", headers, nil)
				},
				func(t *testing.T, w *httptest.ResponseRecorder, title string) {
					assert.Equal(t, http.StatusForbidden, w.Result().StatusCode, title)
					assert.Empty(t, w.Header().Get("Access-Control-Allow-Credentials"), title)
					assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"), title)
				},
			},
			{
				"cors (POST): allow all origins",
				func() *env.ENV {
					environ := test.NewTestENV(t)
					environ.CORSAllowedOrigins = []string{}
					return environ
				},
				func(r *gin.Engine) {
					r.POST("/", func(ctx *gin.Context) {
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
					return performRequest(t, r, http.MethodPost, "/", headers, nil)
				},
				func(t *testing.T, w *httptest.ResponseRecorder, title string) {
					assert.Equal(t, http.StatusNoContent, w.Result().StatusCode, title)
					assert.Equal(t,
						"true",
						w.Result().Header.Get("Access-Control-Allow-Credentials"),
						title,
					)
					assert.Equal(t,
						"*",
						w.Result().Header.Get("Access-Control-Allow-Origin"),
						title,
					)
				},
			},
			{
				"cors (POST): wrong origins ",
				func() *env.ENV {
					environ := test.NewTestENV(t)
					environ.CORSAllowedOrigins = []string{"http://localhost:8000"}
					return environ
				},
				func(r *gin.Engine) {
					r.POST("/", func(ctx *gin.Context) {
						ctx.AbortWithStatus(http.StatusNoContent)
					})
				},
				func(r *gin.Engine) *httptest.ResponseRecorder {
					headers := []header{
						{
							Key:   "Origin",
							Value: "http://localhost:3000",
						},
					}
					return performRequest(t, r, http.MethodPost, "/", headers, nil)
				},
				func(t *testing.T, w *httptest.ResponseRecorder, title string) {
					assert.Equal(t, http.StatusForbidden, w.Result().StatusCode, title)
					assert.Empty(t, w.Header().Get("Access-Control-Allow-Credentials"), title)
					assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"), title)
				},
			},
			{
				"cors (DELETE): allow all origins",
				func() *env.ENV {
					environ := test.NewTestENV(t)
					environ.CORSAllowedOrigins = []string{}
					return environ
				},
				func(r *gin.Engine) {
					r.DELETE("/", func(ctx *gin.Context) {
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
					return performRequest(t, r, http.MethodDelete, "/", headers, nil)
				},
				func(t *testing.T, w *httptest.ResponseRecorder, title string) {
					assert.Equal(t, http.StatusNoContent, w.Result().StatusCode, title)
					assert.Equal(t,
						"true",
						w.Result().Header.Get("Access-Control-Allow-Credentials"),
						title,
					)
					assert.Equal(t,
						"*",
						w.Result().Header.Get("Access-Control-Allow-Origin"),
						title,
					)
				},
			},
			{
				"cors (DELETE): wrong origins ",
				func() *env.ENV {
					environ := test.NewTestENV(t)
					environ.CORSAllowedOrigins = []string{"http://localhost:8000"}
					return environ
				},
				func(r *gin.Engine) {
					r.DELETE("/", func(ctx *gin.Context) {
						ctx.AbortWithStatus(http.StatusNoContent)
					})
				},
				func(r *gin.Engine) *httptest.ResponseRecorder {
					headers := []header{
						{
							Key:   "Origin",
							Value: "http://localhost:3000",
						},
					}
					return performRequest(t, r, http.MethodDelete, "/", headers, nil)
				},
				func(t *testing.T, w *httptest.ResponseRecorder, title string) {
					assert.Equal(t, http.StatusForbidden, w.Result().StatusCode, title)
					assert.Empty(t, w.Header().Get("Access-Control-Allow-Credentials"), title)
					assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"), title)
				},
			},
			{
				"cors (PUT): allow all origins",
				func() *env.ENV {
					environ := test.NewTestENV(t)
					environ.CORSAllowedOrigins = []string{}
					return environ
				},
				func(r *gin.Engine) {
					r.PUT("/", func(ctx *gin.Context) {
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
					return performRequest(t, r, http.MethodPut, "/", headers, nil)
				},
				func(t *testing.T, w *httptest.ResponseRecorder, title string) {
					assert.Equal(t, http.StatusNoContent, w.Result().StatusCode, title)
					assert.Equal(t,
						"true",
						w.Result().Header.Get("Access-Control-Allow-Credentials"),
						title,
					)
					assert.Equal(t,
						"*",
						w.Result().Header.Get("Access-Control-Allow-Origin"),
						title,
					)
				},
			},
			{
				"cors (PUT): wrong origins ",
				func() *env.ENV {
					environ := test.NewTestENV(t)
					environ.CORSAllowedOrigins = []string{"http://localhost:8000"}
					return environ
				},
				func(r *gin.Engine) {
					r.PUT("/", func(ctx *gin.Context) {
						ctx.AbortWithStatus(http.StatusNoContent)
					})
				},
				func(r *gin.Engine) *httptest.ResponseRecorder {
					headers := []header{
						{
							Key:   "Origin",
							Value: "http://localhost:3000",
						},
					}
					return performRequest(t, r, http.MethodPut, "/", headers, nil)
				},
				func(t *testing.T, w *httptest.ResponseRecorder, title string) {
					assert.Equal(t, http.StatusForbidden, w.Result().StatusCode, title)
					assert.Empty(t, w.Header().Get("Access-Control-Allow-Credentials"), title)
					assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"), title)
				},
			},
			{
				"cors (HEAD): allow all origins",
				func() *env.ENV {
					environ := test.NewTestENV(t)
					environ.CORSAllowedOrigins = []string{}
					return environ
				},
				func(r *gin.Engine) {
					r.HEAD("/", func(ctx *gin.Context) {
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
					return performRequest(t, r, http.MethodHead, "/", headers, nil)
				},
				func(t *testing.T, w *httptest.ResponseRecorder, title string) {
					assert.Equal(t, http.StatusNoContent, w.Result().StatusCode, title)
					assert.Equal(t,
						"true",
						w.Result().Header.Get("Access-Control-Allow-Credentials"),
						title,
					)
					assert.Equal(t,
						"*",
						w.Result().Header.Get("Access-Control-Allow-Origin"),
						title,
					)
				},
			},
			{
				"cors (HEAD): wrong origins ",
				func() *env.ENV {
					environ := test.NewTestENV(t)
					environ.CORSAllowedOrigins = []string{"http://localhost:8000"}
					return environ
				},
				func(r *gin.Engine) {
					r.HEAD("/", func(ctx *gin.Context) {
						ctx.AbortWithStatus(http.StatusNoContent)
					})
				},
				func(r *gin.Engine) *httptest.ResponseRecorder {
					headers := []header{
						{
							Key:   "Origin",
							Value: "http://localhost:3000",
						},
					}
					return performRequest(t, r, http.MethodHead, "/", headers, nil)
				},
				func(t *testing.T, w *httptest.ResponseRecorder, title string) {
					assert.Equal(t, http.StatusForbidden, w.Result().StatusCode, title)
					assert.Empty(t, w.Header().Get("Access-Control-Allow-Credentials"), title)
					assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"), title)
				},
			},
			{
				"cors (PATCH): allow all origins",
				func() *env.ENV {
					environ := test.NewTestENV(t)
					environ.CORSAllowedOrigins = []string{}
					return environ
				},
				func(r *gin.Engine) {
					r.PATCH("/", func(ctx *gin.Context) {
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
					return performRequest(t, r, http.MethodPatch, "/", headers, nil)
				},
				func(t *testing.T, w *httptest.ResponseRecorder, title string) {
					assert.Equal(t, http.StatusNoContent, w.Result().StatusCode, title)
					assert.Equal(t,
						"true",
						w.Result().Header.Get("Access-Control-Allow-Credentials"),
						title,
					)
					assert.Equal(t,
						"*",
						w.Result().Header.Get("Access-Control-Allow-Origin"),
						title,
					)
				},
			},
			{
				"cors (PATCH): wrong origins ",
				func() *env.ENV {
					environ := test.NewTestENV(t)
					environ.CORSAllowedOrigins = []string{"http://localhost:8000"}
					return environ
				},
				func(r *gin.Engine) {
					r.PATCH("/", func(ctx *gin.Context) {
						ctx.AbortWithStatus(http.StatusNoContent)
					})
				},
				func(r *gin.Engine) *httptest.ResponseRecorder {
					headers := []header{
						{
							Key:   "Origin",
							Value: "http://localhost:3000",
						},
					}
					return performRequest(t, r, http.MethodPatch, "/", headers, nil)
				},
				func(t *testing.T, w *httptest.ResponseRecorder, title string) {
					assert.Equal(t, http.StatusForbidden, w.Result().StatusCode, title)
					assert.Empty(t, w.Header().Get("Access-Control-Allow-Credentials"), title)
					assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"), title)
				},
			},
		}

		for _, tt := range tests {
			router := gin.New()
			router.Use(CORS(tt.getEnvFn()))

			tt.setupFn(router)
			tt.testFn(t, tt.runFn(router), tt.title)
		}
	})
}
