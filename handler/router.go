package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/nathanbizkit/article-management/middleware"
)

const APIGroupPath = "/api/v1"

// Route links handlers to http api router
func Route(router *gin.Engine, h *Handler) {
	root := router.Group(APIGroupPath)
	{
		public := root.Group("")

		public.POST("/login", h.Login)
		public.POST("/register", h.Register)
		public.POST("/refresh_token", h.RefreshToken)

	}

	{
		unsecured := root.Group("")

		strictCookie := false
		unsecured.Use(middleware.Auth(h.logger, h.auth, strictCookie))

		unsecured.GET("/articles", h.GetArticles)
		unsecured.GET("/articles/:slug", h.GetArticle)
		unsecured.GET("/articles/:slug/comments", h.GetComments)

		unsecured.GET("/tags", h.GetTags)
	}

	{
		private := root.Group("")

		strictCookie := true
		private.Use(middleware.Auth(h.logger, h.auth, strictCookie))

		private.GET("/me", h.GetCurrentUser)
		private.PUT("/me", h.UpdateCurrentUser)

		private.GET("/profiles/:username", h.ShowProfile)
		private.POST("/profiles/:username/follow", h.FollowUser)
		private.DELETE("/profiles/:username/follow", h.UnfollowUser)

		private.GET("/articles/feed", h.GetFeedArticles)
		private.POST("/articles", h.CreateArticle)
		private.PUT("/articles/:slug", h.UpdateArticle)
		private.DELETE("/articles/:slug", h.DeleteArticle)

		private.POST("/articles/:slug/comments", h.CreateComment)
		private.DELETE("/articles/:slug/comments/:id", h.DeleteComment)

		private.POST("/articles/:slug/favorite", h.FavoriteArticle)
		private.DELETE("/articles/:slug/favorite", h.UnfavoriteArticle)
	}
}
