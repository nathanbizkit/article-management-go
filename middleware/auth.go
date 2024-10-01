package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nathanbizkit/article-management/auth"
)

// Auth guards against unauthenticated incoming request
func Auth(a *auth.Auth) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := a.GetUserID(ctx)
		if err != nil {
			ctx.AbortWithStatusJSON(
				http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
			return
		}

		a.SetContextUserID(ctx, id)
	}
}
