package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nathanbizkit/article-management/auth"
	"github.com/rs/zerolog"
)

// Auth guards against unauthenticated incoming request
func Auth(l *zerolog.Logger, a *auth.Auth, strictCookie bool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		refresh := false
		id, err := a.GetUserID(ctx, strictCookie, refresh)
		if err != nil {
			msg := "unauthenticated"
			err = fmt.Errorf("unauthenticated: %w", err)
			l.Error().Err(err).Msg(msg)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": msg})
			return
		}

		a.SetContextUserID(ctx, id)
	}
}
