package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nathanbizkit/article-management/message"
	"github.com/nathanbizkit/article-management/model"
)

func (h *Handler) CreateComment(ctx *gin.Context) {
	h.logger.Info().Interface("req", ctx.Request).Msg("create comment")

	userID := h.auth.GetContextUserID(ctx)

	currentUser, err := h.us.GetByID(ctx.Request.Context(), userID)
	if err != nil {
		h.logger.Error().Err(err).
			Msg(fmt.Sprintf("current user (id=%d) not found", userID))
		ctx.AbortWithStatusJSON(http.StatusNotFound,
			gin.H{"error": "current user not found"})
		return
	}

	articleID, err := strconv.Atoi(ctx.Param("slug"))
	if err != nil {
		h.logger.Error().Err(err).
			Msg(fmt.Sprintf("cannot convert slug (%s) into integer", ctx.Param("slug")))
		ctx.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{"error": "invalid article id"})
		return
	}

	article, err := h.as.GetByID(ctx.Request.Context(), uint(articleID))
	if err != nil {
		h.logger.Error().Err(err).
			Msg(fmt.Sprintf("requested article (slug=%d) not found", articleID))
		ctx.AbortWithStatusJSON(http.StatusNotFound,
			gin.H{"error": "requested article not found"})
		return
	}

	var r message.CreateCommentRequest
	err = ctx.ShouldBindJSON(&r)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to bind request body")
		ctx.AbortWithStatusJSON(http.StatusBadRequest,
			gin.H{"error": "invalid request body"})
		return
	}

	comment := model.Comment{
		Body:      r.Body,
		UserID:    userID,
		Author:    *currentUser,
		ArticleID: article.ID,
		Article:   *article,
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
		ctx.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": msg})
		return
	}

	ctx.JSON(http.StatusOK, message.CommentResponse{
		ID:   createdComment.ID,
		Body: createdComment.Body,
		Author: message.ProfileResponse{
			Username:  currentUser.Username,
			Bio:       currentUser.Bio,
			Image:     currentUser.Image,
			Following: false,
		},
		CreatedAt: createdComment.CreatedAt,
		UpdatedAt: createdComment.UpdatedAt,
	})
}

func (h *Handler) GetComments(ctx *gin.Context) {
	// TODO
}

func (h *Handler) DeleteComment(ctx *gin.Context) {
	// TODO
}
