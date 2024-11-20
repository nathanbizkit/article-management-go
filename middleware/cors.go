package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/nathanbizkit/article-management/env"
)

// CORS attaches cors middleware to http engine
func CORS(environ *env.ENV) gin.HandlerFunc {
	config := cors.DefaultConfig()
	config.AllowCredentials = true
	config.AllowAllOrigins = true
	config.AllowWildcard = true
	config.AllowWebSockets = true
	config.AllowBrowserExtensions = true

	if len(environ.CORSAllowedOrigins) != 0 {
		config.AllowAllOrigins = false
		config.AllowOrigins = environ.CORSAllowedOrigins
	}

	return cors.New(config)
}
