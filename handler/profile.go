package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ShowProfile gets a user profile
func (h *Handler) ShowProfile(ctx *gin.Context) {
	h.logger.Info().Msg("show profile")

	currentUser, err := h.GetCurrentUserFromContext(ctx)
	if err != nil {
		msg := "current user not found"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": msg})
		return
	}

	username := ctx.Param("username")
	user, err := h.us.GetByUsername(ctx.Request.Context(), username)
	if err != nil {
		msg := "user not found"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": msg})
		return
	}

	following, err := h.us.IsFollowing(ctx.Request.Context(), currentUser, user)
	if err != nil {
		msg := "failed to get following status"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}

	ctx.AbortWithStatusJSON(http.StatusOK, user.ResponseProfile(following))
}

// FollowUser follows a user
func (h *Handler) FollowUser(ctx *gin.Context) {
	h.logger.Info().Msg("follow user")

	currentUser, err := h.GetCurrentUserFromContext(ctx)
	if err != nil {
		msg := "current user not found"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": msg})
		return
	}

	username := ctx.Param("username")

	if currentUser.Username == username {
		msg := "cannot follow yourself"
		err := fmt.Errorf("user (username: %s) cannot follow yourself", currentUser.Username)
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	user, err := h.us.GetByUsername(ctx.Request.Context(), username)
	if err != nil {
		msg := "user not found"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": msg})
		return
	}

	following, err := h.us.IsFollowing(ctx.Request.Context(), currentUser, user)
	if err != nil {
		msg := "failed to get following status"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}

	if following {
		err := fmt.Errorf("current user (ID: %d) is already following user (ID: %d)", currentUser.ID, user.ID)
		h.logger.Error().Err(err).Msg("current user is already following the user")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "you are already following this user"})
		return
	}

	err = h.us.Follow(ctx.Request.Context(), currentUser, user)
	if err != nil {
		h.logger.Error().Err(err).
			Msg(fmt.Sprintf("failed to follow user: (ID: %d) -> (ID: %d)", currentUser.ID, user.ID))
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to follow user"})
		return
	}

	following = true
	ctx.AbortWithStatusJSON(http.StatusOK, user.ResponseProfile(following))
}

// UnfollowUser unfollows a user
func (h *Handler) UnfollowUser(ctx *gin.Context) {
	h.logger.Info().Msg("unfollow user")

	currentUser, err := h.GetCurrentUserFromContext(ctx)
	if err != nil {
		msg := "current user not found"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": msg})
		return
	}

	username := ctx.Param("username")

	if currentUser.Username == username {
		msg := "cannot unfollow yourself"
		err := fmt.Errorf("user (username: %s) cannot unfollow yourself", currentUser.Username)
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	user, err := h.us.GetByUsername(ctx.Request.Context(), username)
	if err != nil {
		msg := "user not found"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": msg})
		return
	}

	following, err := h.us.IsFollowing(ctx.Request.Context(), currentUser, user)
	if err != nil {
		msg := "failed to get following status"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}

	if !following {
		err := fmt.Errorf("current user (ID: %d) is not following user (ID: %d)", currentUser.ID, user.ID)
		h.logger.Error().Err(err).Msg("current user is not following the user")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "you are not following this user"})
		return
	}

	err = h.us.Unfollow(ctx.Request.Context(), currentUser, user)
	if err != nil {
		h.logger.Error().Err(err).
			Msg(fmt.Sprintf("failed to unfollow user: (ID: %d) -> (ID: %d)", currentUser.ID, user.ID))
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to unfollow user"})
		return
	}

	following = false
	ctx.AbortWithStatusJSON(http.StatusOK, user.ResponseProfile(following))
}
