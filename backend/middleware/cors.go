package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/nathanbizkit/article-management/env"
)

func CORS(e env.ENVer) gin.HandlerFunc {
	config := cors.DefaultConfig()
	config.AllowCredentials = true

	allowedOrigins := e.Values().CorsAllowedOrigins()
	if len(allowedOrigins) > 0 {
		config.AllowAllOrigins = false
		config.AllowOrigins = allowedOrigins
	} else {
		config.AllowAllOrigins = true
	}

	return cors.New(config)
}
