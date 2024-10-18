package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nathanbizkit/article-management/message"
	"github.com/nathanbizkit/article-management/model"
)

// Login logs an existing user in and attaches tokens to cookie
func (h *Handler) Login(ctx *gin.Context) {
	h.logger.Info().Interface("req", ctx.Request).Msg("login")

	var r message.LoginUserRequest
	err := ctx.ShouldBindJSON(&r)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to bind request body")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	user, err := h.us.GetByEmail(ctx.Request.Context(), r.Email)
	if err != nil {
		msg := "user not found"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": msg})
		return
	}

	if !user.CheckPassword(r.Password) {
		msg := "invalid password"
		err := fmt.Errorf("password (%s) is not matched", r.Password)
		h.logger.Error().Err(err).Msgf(msg)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	token, err := h.auth.GenerateToken(user.ID)
	if err != nil {
		msg := "failed to generate token"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}

	h.auth.SetCookieToken(ctx, *token, APIGroupPath)

	ctx.AbortWithStatus(http.StatusOK)
}

// Register creates a new user and attaches tokens to cookie
func (h *Handler) Register(ctx *gin.Context) {
	h.logger.Info().Interface("req", ctx.Request).Msg("register")

	var r message.CreateUserRequest
	err := ctx.ShouldBindJSON(&r)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to bind request body")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	user := model.User{
		Username: r.Username,
		Email:    r.Email,
		Password: r.Password,
		Name:     r.Name,
	}

	err = user.Validate()
	if err != nil {
		h.logger.Error().Err(err).Msg("validation error")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = user.HashPassword()
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to hash password")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid password"})
		return
	}

	createdUser, err := h.us.Create(ctx.Request.Context(), &user)
	if err != nil {
		msg := "failed to create user"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}

	token, err := h.auth.GenerateToken(createdUser.ID)
	if err != nil {
		msg := "failed to generate token"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}

	h.auth.SetCookieToken(ctx, *token, APIGroupPath)

	following := false
	ctx.AbortWithStatusJSON(http.StatusOK, createdUser.ResponseProfile(following))
}

// RefreshToken verifies and renew tokens to cookie
func (h *Handler) RefreshToken(ctx *gin.Context) {
	h.logger.Info().Interface("req", ctx.Request).Msg("refresh token")

	refresh := true
	id, err := h.auth.GetUserID(ctx, refresh)
	if err != nil {
		msg := "failed to extract token from cookie"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}

	token, err := h.auth.GenerateToken(id)
	if err != nil {
		msg := "failed to generate token"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}

	h.auth.SetCookieToken(ctx, *token, APIGroupPath)

	ctx.AbortWithStatus(http.StatusOK)
}

// GetCurrentUser gets current user's profile
func (h *Handler) GetCurrentUser(ctx *gin.Context) {
	h.logger.Info().Interface("req", ctx.Request).Msg("get current user")

	currentUser := h.GetCurrentUserOrAbort(ctx)

	token, err := h.auth.GenerateToken(currentUser.ID)
	if err != nil {
		msg := "failed to generate token"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}

	h.auth.SetCookieToken(ctx, *token, APIGroupPath)

	following := false
	ctx.AbortWithStatusJSON(http.StatusOK, currentUser.ResponseProfile(following))
}

// UpdateCurrentUser updates current user's profile
func (h *Handler) UpdateCurrentUser(ctx *gin.Context) {
	h.logger.Info().Interface("req", ctx.Request).Msg("update current user")

	var r message.UpdateUserRequest
	err := ctx.ShouldBindJSON(&r)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to bind request body")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	currentUser := h.GetCurrentUserOrAbort(ctx)

	isPlainPassword := currentUser.Overwrite(r.Username, r.Email, r.Password, r.Name, r.Bio, r.Image)

	err = currentUser.Validate()
	if err != nil {
		h.logger.Error().Err(err).Msg("validation error")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if isPlainPassword {
		err = currentUser.HashPassword()
		if err != nil {
			h.logger.Error().Err(err).Msg("failed to hash password")
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid password"})
			return
		}
	}

	updatedUser, err := h.us.Update(ctx.Request.Context(), currentUser)
	if err != nil {
		msg := "failed to update profile"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}

	token, err := h.auth.GenerateToken(updatedUser.ID)
	if err != nil {
		msg := "failed to generate token"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}

	h.auth.SetCookieToken(ctx, *token, APIGroupPath)

	following := false
	ctx.AbortWithStatusJSON(http.StatusOK, updatedUser.ResponseProfile(following))
}
