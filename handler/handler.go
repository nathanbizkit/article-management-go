package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/nathanbizkit/article-management/auth"
	"github.com/nathanbizkit/article-management/env"
	"github.com/nathanbizkit/article-management/middleware"
	"github.com/nathanbizkit/article-management/store"
	"github.com/rs/zerolog"
)

// Handler definition
type Handler struct {
	logger *zerolog.Logger
	env    *env.ENV
	auth   *auth.Auth
	us     *store.UserStore
	as     *store.ArticleStore
}

// New returns a new handler with logger, env, auth and database
func New(l *zerolog.Logger, e *env.ENV, auth *auth.Auth, us *store.UserStore, as *store.ArticleStore) *Handler {
	return &Handler{logger: l, env: e, auth: auth, us: us, as: as}
}

// Route links handlers to http api router
func Route(router *gin.Engine, auth *auth.Auth, h *Handler) {
	root := router.Group("/api")
	{
		public := root.Group("")

		public.POST("/login", h.Login)
		public.POST("/register", h.Register)
		public.POST("/refresh_token", h.RefreshToken)
	}

	{
		private := root.Group("")
		private.Use(middleware.Auth(auth))

		private.GET("/me", h.CurrentUser)
		private.PUT("/me", h.UpdateCurrentUser)

		private.GET("/profiles/:username", h.ShowProfile)
		private.POST("/profiles/:username/follow", h.FollowUser)
		private.DELETE("/profiles/:username/follow", h.UnfollowUser)

		private.GET("/articles/feed", h.GetFeedArticles)
		private.GET("/articles", h.GetArticles)
		private.POST("/articles", h.CreateArticle)
		private.GET("/articles/:slug", h.GetArticle)
		private.PUT("/articles/:slug", h.UpdateArticle)
		private.DELETE("/articles/:slug", h.DeleteArticle)

		private.GET("/articles/:slug/comments", h.GetComments)
		private.POST("/articles/:slug/comments", h.CreateComment)
		private.DELETE("/articles/:slug/comments/:id", h.DeleteComment)

		private.POST("/articles/:slug/favorite", h.FavoriteArticle)
		private.DELETE("/articles/:slug/favorite", h.UnfavoriteArticle)

		private.GET("/tags", h.GetTags)
	}
}