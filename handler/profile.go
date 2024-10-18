package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ShowProfile gets a user profile
func (h *Handler) ShowProfile(ctx *gin.Context) {
	h.logger.Info().Interface("req", ctx.Request).Msg("show profile")

	currentUser := h.GetCurrentUserOrAbort(ctx)

	requestUser, err := h.us.GetByUsername(ctx.Request.Context(), ctx.Param("username"))
	if err != nil {
		msg := "user not found"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": msg})
		return
	}

	following, err := h.us.IsFollowing(ctx.Request.Context(), currentUser, requestUser)
	if err != nil {
		msg := "failed to get following status"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}

	ctx.JSON(http.StatusOK, requestUser.ResponseProfile(following))
}

// FollowUser follows a user
func (h *Handler) FollowUser(ctx *gin.Context) {
	h.logger.Info().Interface("req", ctx.Request).Msg("follow user")

	currentUser := h.GetCurrentUserOrAbort(ctx)

	if currentUser.Username == ctx.Param("username") {
		msg := "cannot follow yourself"
		err := fmt.Errorf("user (username: %s) cannot follow yourself", currentUser.Username)
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	requestUser, err := h.us.GetByUsername(ctx.Request.Context(), ctx.Param("username"))
	if err != nil {
		msg := "user not found"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": msg})
		return
	}

	err = h.us.Follow(ctx.Request.Context(), currentUser, requestUser)
	if err != nil {
		h.logger.Error().Err(err).Msg(fmt.Sprintf("failed to follow user: (ID: %d) -> (ID: %d)",
			currentUser.ID, requestUser.ID))
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to follow user"})
		return
	}

	following := true
	ctx.JSON(http.StatusOK, requestUser.ResponseProfile(following))
}

// UnfollowUser unfollows a user
func (h *Handler) UnfollowUser(ctx *gin.Context) {
	h.logger.Info().Interface("req", ctx.Request).Msg("unfollow user")

	currentUser := h.GetCurrentUserOrAbort(ctx)

	if currentUser.Username == ctx.Param("username") {
		msg := "cannot unfollow yourself"
		err := fmt.Errorf("user (username: %s) cannot unfollow yourself", currentUser.Username)
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	requestUser, err := h.us.GetByUsername(ctx.Request.Context(), ctx.Param("username"))
	if err != nil {
		msg := "user not found"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": msg})
		return
	}

	following, err := h.us.IsFollowing(ctx.Request.Context(), currentUser, requestUser)
	if err != nil {
		msg := "failed to get following status"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}

	if !following {
		err := fmt.Errorf("current user (ID: %d) is not following request user (ID: %d)",
			currentUser.ID, requestUser.ID)
		h.logger.Error().Err(err).Msg("current user is not following request user")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "you are not following the user"})
		return
	}

	err = h.us.Unfollow(ctx.Request.Context(), currentUser, requestUser)
	if err != nil {
		h.logger.Error().Err(err).Msg(fmt.Sprintf("failed to unfollow user: (ID: %d) -> (ID: %d)",
			currentUser.ID, requestUser.ID))
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to unfollow user"})
		return
	}

	following = false
	ctx.JSON(http.StatusOK, requestUser.ResponseProfile(following))
}
