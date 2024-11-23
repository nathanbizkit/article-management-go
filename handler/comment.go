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

	currentUser, err := h.GetCurrentUserFromContext(ctx)
	if err != nil {
		msg := "current user not found"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": msg})
		return
	}

	slug, err := h.GetIDFromParam(ctx, "slug")
	if err != nil {
		msg := "invalid slug"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	article, err := h.as.GetByID(ctx.Request.Context(), slug)
	if err != nil {
		h.logger.Error().Err(err).Msg(fmt.Sprintf("article (slug=%d) not found", slug))
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "article not found"})
		return
	}

	var req message.CreateCommentRequest
	err = ctx.ShouldBindJSON(&req)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to bind request body")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	comment := model.Comment{
		Body:      req.Body,
		UserID:    currentUser.ID,
		Author:    *currentUser,
		ArticleID: article.ID,
	}

	err = comment.Validate()
	if err != nil {
		err := fmt.Errorf("validation error: %w", err)
		h.logger.Error().Err(err).Msg("validation error")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createdComment, err := h.as.CreateComment(ctx.Request.Context(), &comment)
	if err != nil {
		msg := "failed to create comment"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}

	following := false
	ctx.AbortWithStatusJSON(http.StatusOK, createdComment.ResponseComment(following))
}

// GetComments gets comments of an article
func (h *Handler) GetComments(ctx *gin.Context) {
	h.logger.Info().Msg("get comments")

	slug, err := h.GetIDFromParam(ctx, "slug")
	if err != nil {
		msg := "invalid slug"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	article, err := h.as.GetByID(ctx.Request.Context(), slug)
	if err != nil {
		h.logger.Error().Err(err).Msg(fmt.Sprintf("article (slug=%d) not found", slug))
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

	resp := make([]message.CommentResponse, 0, len(comments))
	for _, c := range comments {
		following, err := h.us.IsFollowing(ctx.Request.Context(), currentUser, &c.Author)
		if err != nil {
			msg := "failed to get following status"
			h.logger.Error().Err(err).Msg(msg)
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		resp = append(resp, c.ResponseComment(following))
	}

	ctx.AbortWithStatusJSON(http.StatusOK, message.CommentsResponse{Comments: resp})
}

// DeleteComment deletes a comment from an article
func (h *Handler) DeleteComment(ctx *gin.Context) {
	h.logger.Info().Msg("delete comment")

	currentUser, err := h.GetCurrentUserFromContext(ctx)
	if err != nil {
		msg := "current user not found"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": msg})
		return
	}

	slug, err := h.GetIDFromParam(ctx, "slug")
	if err != nil {
		msg := "invalid slug"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	id, err := h.GetIDFromParam(ctx, "id")
	if err != nil {
		msg := "invalid comment id"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	comment, err := h.as.GetCommentByID(ctx.Request.Context(), id)
	if err != nil {
		h.logger.Error().Err(err).Msg(fmt.Sprintf("comment (id=%d) not found", id))
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "comment not found"})
		return
	}

	if slug != comment.ArticleID {
		msg := "the comment is not from this article"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	if comment.UserID != currentUser.ID {
		err := fmt.Errorf(
			"current user (id=%d) is forbidden to delete this comment (id=%d)",
			currentUser.ID, comment.ID,
		)
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

	ctx.AbortWithStatus(http.StatusNoContent)
}
