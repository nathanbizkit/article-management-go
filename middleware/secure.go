package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/nathanbizkit/article-management/env"
	"github.com/unrolled/secure"
)

// Secure attaches secure tls middleware to http engine
func Secure(e *env.ENV) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		isDevelopment := e.AppMode == "dev" || e.AppMode == "develop" ||
			e.AppMode == "test" || e.AppMode == "testing"

		opts := secure.Options{
			SSLRedirect:   true,
			IsDevelopment: isDevelopment,
		}

		if isDevelopment {
			opts.SSLHost = fmt.Sprintf("localhost:%s", e.AppTLSPort)
		}

		secureMiddleware := secure.New(opts)

		err := secureMiddleware.Process(ctx.Writer, ctx.Request)
		if err != nil {
			ctx.Abort()
			return
		}

		// avoid header rewrite if response is a redirection
		status := ctx.Writer.Status()
		if status > 300 && status < 399 {
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
