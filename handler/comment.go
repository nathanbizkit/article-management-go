package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nathanbizkit/article-management/message"
	"github.com/nathanbizkit/article-management/model"
)

// CreateComment creates a comment
func (h *Handler) CreateComment(ctx *gin.Context) {
	h.logger.Info().Msg("create comment")

	currentUser := h.GetCurrentUserOrAbort(ctx)

	articleID := h.GetParamAsIDOrAbort(ctx, "slug")
	article, err := h.as.GetByID(ctx.Request.Context(), uint(articleID))
	if err != nil {
		h.logger.Error().Err(err).Msg(fmt.Sprintf("article (slug=%d) not found", articleID))
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "article not found"})
		return
	}

	var r message.CreateCommentRequest
	err = ctx.ShouldBindJSON(&r)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to bind request body")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	comment := model.Comment{
		Body:      r.Body,
		UserID:    currentUser.ID,
		Author:    *currentUser,
		ArticleID: article.ID,
	}

	err = comment.Validate()
	if err != nil {
		err := fmt.Errorf("validation error: %w", err)
		h.logger.Error().Err(err).Msg("validation error")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	createdComment, err := h.as.CreateComment(ctx.Request.Context(), &comment)
	if err != nil {
		msg := "failed to create comment"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}

	ctx.JSON(http.StatusOK, createdComment.ResponseComment(false))
}

// GetComments gets comments of an article
func (h *Handler) GetComments(ctx *gin.Context) {
	h.logger.Info().Msg("get comments")

	currentUser := h.GetCurrentUserOrAbort(ctx)

	articleID := h.GetParamAsIDOrAbort(ctx, "slug")
	article, err := h.as.GetByID(ctx.Request.Context(), uint(articleID))
	if err != nil {
		h.logger.Error().Err(err).Msg(fmt.Sprintf("article (slug=%d) not found", articleID))
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "article not found"})
		return
	}

	comments, err := h.as.GetComments(ctx.Request.Context(), article)
	if err != nil {
		msg := "failed to get comments"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}

	crs := make([]message.CommentResponse, 0, len(comments))
	for _, c := range comments {
		following, err := h.us.IsFollowing(ctx.Request.Context(), currentUser, &c.Author)
		if err != nil {
			msg := "failed to get following status"
			h.logger.Error().Err(err).Msg(msg)
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		crs = append(crs, c.ResponseComment(following))
	}

	ctx.JSON(http.StatusOK, message.CommentsResponse{Comments: crs})
}

// DeleteComment deletes a comment from an article
func (h *Handler) DeleteComment(ctx *gin.Context) {
	h.logger.Info().Msg("delete comment")

	currentUser := h.GetCurrentUserOrAbort(ctx)

	commentID := h.GetParamAsIDOrAbort(ctx, "id")
	comment, err := h.as.GetCommentByID(ctx.Request.Context(), commentID)
	if err != nil {
		h.logger.Error().Err(err).Msg(fmt.Sprintf("comment (id=%d) not found", commentID))
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "comment not found"})
		return
	}

	if ctx.Param("slug") != fmt.Sprintf("%d", comment.ArticleID) {
		msg := "the comment is not in the article"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	if comment.UserID != currentUser.ID {
		err := fmt.Errorf("current user (id=%d) is forbidden to delete this comment (id=%d)",
			currentUser.ID, comment.ID)
		msg := "forbidden"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": msg})
		return
	}

	err = h.as.DeleteComment(ctx.Request.Context(), comment)
	if err != nil {
		msg := "failed to delete comment"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}

	ctx.AbortWithStatus(http.StatusOK)
}
