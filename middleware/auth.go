package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nathanbizkit/article-management-go/auth"
	"github.com/rs/zerolog"
)

// Auth guards against unauthorized incoming request
func Auth(l *zerolog.Logger, authen *auth.Auth, strictCookie bool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		refresh := false
		id, err := authen.GetUserID(ctx, strictCookie, refresh)
		if err != nil {
			msg := "unauthorized"
			err = fmt.Errorf("unauthorized: %w", err)
			l.Error().Err(err).Msg(msg)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": msg})
			return
		}

		authen.SetContextUserID(ctx, id)

		ctx.Next()
	}
}
