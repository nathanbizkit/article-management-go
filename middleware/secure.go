package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nathanbizkit/article-management/env"
	"github.com/unrolled/secure"
)

// Secure attaches secure tls middleware to http engine
func Secure(e *env.ENV) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		isDevelopment := e.AppMode == "dev" || e.AppMode == "develop" ||
			e.AppMode == "test" || e.AppMode == "testing"

		secureMiddleware := secure.New(secure.Options{
			SSLRedirect:   true,
			SSLHost:       "localhost:8443",
			IsDevelopment: isDevelopment,
		})

		err := secureMiddleware.Process(ctx.Writer, ctx.Request)
		if err != nil {
			msg := "failed to process tls connection"
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		ctx.Next()
	}
}
