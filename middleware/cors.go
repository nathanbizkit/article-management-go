package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/nathanbizkit/article-management/env"
)

// CORS attaches cors middleware to http engine
func CORS(e *env.ENV) gin.HandlerFunc {
	config := cors.DefaultConfig()
	config.AllowCredentials = true

	if len(e.CORSAllowedOrigins) > 0 {
		config.AllowAllOrigins = false
		config.AllowOrigins = e.CORSAllowedOrigins
	} else {
		config.AllowAllOrigins = true
	}

	return cors.New(config)
}
