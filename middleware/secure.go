package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/nathanbizkit/article-management/env"
	"github.com/unrolled/secure"
)

// Secure attaches secure tls middleware to http engine
func Secure(environ *env.ENV) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		opts := secure.Options{
			SSLRedirect:   true,
			IsDevelopment: environ.IsDevelopment,
		}

		if environ.IsDevelopment {
			opts.SSLHost = fmt.Sprintf("localhost:%s", environ.AppTLSPort)
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
