package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nathanbizkit/article-management/message"
	"github.com/nathanbizkit/article-management/model"
)

const (
	defaultLimit  = 20
	defaultOffset = 0
)

// CreateArticle creates an article
func (h *Handler) CreateArticle(ctx *gin.Context) {
	h.logger.Info().Msg("create article")

	currentUser := h.GetCurrentUserOrAbort(ctx)

	var req message.CreateArticleRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to bind request body")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	tags := make([]model.Tag, 0, len(req.Tags))
	for _, t := range req.Tags {
		tags = append(tags, model.Tag{Name: t})
	}

	article := model.Article{
		Title:       req.Title,
		Description: req.Description,
		Body:        req.Body,
		UserID:      currentUser.ID,
		Author:      *currentUser,
		Tags:        tags,
	}

	err = article.Validate()
	if err != nil {
		h.logger.Error().Err(err).Msg("validation error")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createdArticle, err := h.as.Create(ctx.Request.Context(), &article)
	if err != nil {
		msg := "failed to create article"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}

	favorited := false
	following := false
	ctx.JSON(http.StatusOK, createdArticle.ResponseArticle(favorited, following))
}

// GetArticle gets an article
func (h *Handler) GetArticle(ctx *gin.Context) {
	h.logger.Info().Msg("get article")

	articleID := h.GetParamAsIDOrAbort(ctx, "slug")
	article, err := h.as.GetByID(ctx.Request.Context(), articleID)
	if err != nil {
		h.logger.Error().Err(err).Msg(fmt.Sprintf("article (slug=%d) not found", articleID))
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "article not found"})
		return
	}

	var currentUser *model.User

	userID := h.authen.GetContextUserID(ctx)
	if userID != 0 {
		currentUser, err = h.us.GetByID(ctx.Request.Context(), userID)
		if err != nil {
			h.logger.Error().Err(err).Msg(fmt.Sprintf("current user (id=%d) not found", userID))
			ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "current user not found"})
			return
		}
	}

	favorited, err := h.as.IsFavorited(ctx.Request.Context(), article, currentUser)
	if err != nil {
		msg := "failed to get favorited status"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}

	following, err := h.us.IsFollowing(ctx.Request.Context(), currentUser, &article.Author)
	if err != nil {
		msg := "failed to get following status"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}

	ctx.JSON(http.StatusOK, article.ResponseArticle(favorited, following))
}

// GetArticles gets recent articles globally
func (h *Handler) GetArticles(ctx *gin.Context) {
	h.logger.Info().Msg("get articles")

	var favoritedBy *model.User
	favByUsername := ctx.Query("favorited")
	if favByUsername != "" {
		var err error
		favoritedBy, err = h.us.GetByUsername(ctx.Request.Context(), favByUsername)
		if err != nil {
			favoritedBy = nil
			h.logger.Warn().Msg("skipped: cannot find user (favorited by)")
		}
	}

	tagName := ctx.Query("tag")
	author := ctx.Query("username")
	limit, offset := h.GetPaginationQuery(ctx, defaultLimit, defaultOffset)

	articles, err := h.as.GetArticles(ctx.Request.Context(), tagName, author, favoritedBy, limit, offset)
	if err != nil {
		msg := "failed to search articles"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}

	var currentUser *model.User

	userID := h.authen.GetContextUserID(ctx)
	if userID != 0 {
		currentUser, err = h.us.GetByID(ctx.Request.Context(), userID)
		if err != nil {
			h.logger.Error().Err(err).Msg(fmt.Sprintf("current user (id=%d) not found", userID))
			ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "current user not found"})
			return
		}
	}

	resp := make([]message.ArticleResponse, 0, len(articles))
	for _, article := range articles {
		favorited, err := h.as.IsFavorited(ctx.Request.Context(), &article, currentUser)
		if err != nil {
			msg := "failed to get favorited status"
			h.logger.Error().Err(err).Msg(msg)
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		following, err := h.us.IsFollowing(ctx.Request.Context(), currentUser, &article.Author)
		if err != nil {
			msg := "failed to get following status"
			h.logger.Error().Err(err).Msg(msg)
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		resp = append(resp, article.ResponseArticle(favorited, following))
	}

	ctx.JSON(http.StatusOK, message.ArticlesResponse{Articles: resp, ArticlesCount: int64(len(resp))})
}

// GetFeedArticles gets recent articles from users that current user follows
func (h *Handler) GetFeedArticles(ctx *gin.Context) {
	h.logger.Info().Msg("get feed articles")

	currentUser := h.GetCurrentUserOrAbort(ctx)

	userIDs, err := h.us.GetFollowingUserIDs(ctx.Request.Context(), currentUser)
	if err != nil {
		h.logger.Error().Err(err).Msg(fmt.Sprintf("failed to get following user ids of user %d", currentUser.ID))
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to get followers"})
		return
	}

	limit, offset := h.GetPaginationQuery(ctx, defaultLimit, defaultOffset)

	articles, err := h.as.GetFeedArticles(ctx.Request.Context(), userIDs, limit, offset)
	if err != nil {
		msg := "failed to search articles from user's followers"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}

	following := true
	resp := make([]message.ArticleResponse, 0, len(articles))
	for _, article := range articles {
		favorited, err := h.as.IsFavorited(ctx.Request.Context(), &article, currentUser)
		if err != nil {
			msg := "failed to get favorited status"
			h.logger.Error().Err(err).Msg(msg)
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		resp = append(resp, article.ResponseArticle(favorited, following))
	}

	ctx.JSON(http.StatusOK, message.ArticlesResponse{Articles: resp, ArticlesCount: int64(len(resp))})
}

// UpdateArticle updates an article
func (h *Handler) UpdateArticle(ctx *gin.Context) {
	h.logger.Info().Msg("update article")

	currentUser := h.GetCurrentUserOrAbort(ctx)

	slug := h.GetParamAsIDOrAbort(ctx, "slug")
	article, err := h.as.GetByID(ctx.Request.Context(), slug)
	if err != nil {
		h.logger.Error().Err(err).Msg(fmt.Sprintf("article (slug=%d) not found", slug))
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "article not found"})
		return
	}

	if article.Author.ID != currentUser.ID {
		msg := "forbidden"
		err := fmt.Errorf("user (id=%d) attempted to update user's article (id=%d)", currentUser.ID, slug)
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": msg})
		return
	}

	var req message.UpdateArticleRequest
	err = ctx.ShouldBindJSON(&req)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to bind request body")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	article.Overwrite(req.Title, req.Description, req.Body)

	err = article.Validate()
	if err != nil {
		h.logger.Error().Err(err).Msg("validation error")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedArticle, err := h.as.Update(ctx.Request.Context(), article)
	if err != nil {
		msg := "failed to update article"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}

	favorited := false
	following := false
	ctx.JSON(http.StatusOK, updatedArticle.ResponseArticle(favorited, following))
}

// DeleteArticle deletes an article
func (h *Handler) DeleteArticle(ctx *gin.Context) {
	h.logger.Info().Msg("delete article")

	currentUser := h.GetCurrentUserOrAbort(ctx)

	slug := h.GetParamAsIDOrAbort(ctx, "slug")
	article, err := h.as.GetByID(ctx.Request.Context(), slug)
	if err != nil {
		h.logger.Error().Err(err).Msg(fmt.Sprintf("article (slug=%d) not found", slug))
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "article not found"})
		return
	}

	if article.Author.ID != currentUser.ID {
		msg := "forbidden"
		err := fmt.Errorf("user (id=%d) attempted to delete user's article (id=%d)", currentUser.ID, slug)
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": msg})
		return
	}

	err = h.as.Delete(ctx.Request.Context(), article)
	if err != nil {
		msg := "failed to delete article"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}

	ctx.AbortWithStatus(http.StatusOK)
}

// FavoriteArticle adds an article to user's favorites
func (h *Handler) FavoriteArticle(ctx *gin.Context) {
	h.logger.Info().Msg("favorite article")

	currentUser := h.GetCurrentUserOrAbort(ctx)

	slug := h.GetParamAsIDOrAbort(ctx, "slug")
	article, err := h.as.GetByID(ctx.Request.Context(), slug)
	if err != nil {
		h.logger.Error().Err(err).Msg(fmt.Sprintf("article (slug=%d) not found", slug))
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "article not found"})
		return
	}

	err = h.as.AddFavorite(ctx.Request.Context(), article, currentUser,
		func(favoritesCount int64, updatedAt time.Time) {
			article.FavoritesCount = favoritesCount
			article.UpdatedAt = updatedAt
		})
	if err != nil {
		msg := "failed to favorite the article"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}

	following, err := h.us.IsFollowing(ctx.Request.Context(), currentUser, &article.Author)
	if err != nil {
		msg := "failed to get following status"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}

	favorited := true
	ctx.JSON(http.StatusOK, article.ResponseArticle(favorited, following))
}

// UnfavoriteArticle removes an article from user's favorites
func (h *Handler) UnfavoriteArticle(ctx *gin.Context) {
	h.logger.Info().Msg("unfavorite article")

	currentUser := h.GetCurrentUserOrAbort(ctx)

	slug := h.GetParamAsIDOrAbort(ctx, "slug")
	article, err := h.as.GetByID(ctx.Request.Context(), slug)
	if err != nil {
		h.logger.Error().Err(err).Msg(fmt.Sprintf("article (slug=%d) not found", slug))
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "article not found"})
		return
	}

	err = h.as.DeleteFavorite(ctx.Request.Context(), article, currentUser,
		func(favoritesCount int64, updatedAt time.Time) {
			article.FavoritesCount = favoritesCount
			article.UpdatedAt = updatedAt
		})
	if err != nil {
		msg := "failed to unfavorite the article"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}

	following, err := h.us.IsFollowing(ctx.Request.Context(), currentUser, &article.Author)
	if err != nil {
		msg := "failed to get following status"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}

	favorited := false
	ctx.JSON(http.StatusOK, article.ResponseArticle(favorited, following))
}
