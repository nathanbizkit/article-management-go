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
	h.logger.Info().Msg("login")

	var req message.LoginUserRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to bind request body")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	user, err := h.us.GetByEmail(ctx.Request.Context(), req.Email)
	if err != nil {
		msg := "user not found"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": msg})
		return
	}

	if !user.CheckPassword(req.Password) {
		msg := "invalid password"
		err := fmt.Errorf("password (%s) is not matched", req.Password)
		h.logger.Error().Err(err).Msgf(msg)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	token, err := h.authen.GenerateToken(user.ID)
	if err != nil {
		msg := "failed to generate token"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}

	h.authen.SetCookieToken(ctx, *token, APIGroupPath)

	ctx.AbortWithStatus(http.StatusNoContent)
}

// Register creates a new user and attaches tokens to cookie
func (h *Handler) Register(ctx *gin.Context) {
	h.logger.Info().Msg("register")

	var req message.CreateUserRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to bind request body")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	user := model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
	}

	isPlainPassword := true
	err = user.Validate(isPlainPassword)
	if err != nil {
		err := fmt.Errorf("validation error: %w", err)
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

	token, err := h.authen.GenerateToken(createdUser.ID)
	if err != nil {
		msg := "failed to generate token"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}

	h.authen.SetCookieToken(ctx, *token, APIGroupPath)

	following := false
	ctx.AbortWithStatusJSON(http.StatusOK, createdUser.ResponseProfile(following))
}

// RefreshToken verifies and renew tokens to cookie
func (h *Handler) RefreshToken(ctx *gin.Context) {
	h.logger.Info().Msg("refresh token")

	strictCookie := true
	refresh := true
	id, err := h.authen.GetUserID(ctx, strictCookie, refresh)
	if err != nil {
		msg := "failed to extract token from cookie"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}

	user, err := h.us.GetByID(ctx.Request.Context(), id)
	if err != nil {
		msg := "user not found"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": msg})
		return
	}

	token, err := h.authen.GenerateToken(user.ID)
	if err != nil {
		msg := "failed to generate token"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}

	h.authen.SetCookieToken(ctx, *token, APIGroupPath)

	ctx.AbortWithStatus(http.StatusNoContent)
}

// GetCurrentUser gets current user's profile
func (h *Handler) GetCurrentUser(ctx *gin.Context) {
	h.logger.Info().Msg("get current user")

	currentUser, err := h.GetCurrentUserFromContext(ctx)
	if err != nil {
		msg := "current user not found"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": msg})
		return
	}

	token, err := h.authen.GenerateToken(currentUser.ID)
	if err != nil {
		msg := "failed to generate token"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}

	h.authen.SetCookieToken(ctx, *token, APIGroupPath)

	following := false
	ctx.AbortWithStatusJSON(http.StatusOK, currentUser.ResponseProfile(following))
}

// UpdateCurrentUser updates current user's profile
func (h *Handler) UpdateCurrentUser(ctx *gin.Context) {
	h.logger.Info().Msg("update current user")

	currentUser, err := h.GetCurrentUserFromContext(ctx)
	if err != nil {
		msg := "current user not found"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": msg})
		return
	}

	var req message.UpdateUserRequest
	err = ctx.ShouldBindJSON(&req)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to bind request body")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	isPlainPassword := currentUser.Overwrite(req.Username, req.Email, req.Password, req.Name, req.Bio, req.Image)

	err = currentUser.Validate(isPlainPassword)
	if err != nil {
		err := fmt.Errorf("validation error: %w", err)
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

	token, err := h.authen.GenerateToken(updatedUser.ID)
	if err != nil {
		msg := "failed to generate token"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}

	h.authen.SetCookieToken(ctx, *token, APIGroupPath)

	following := false
	ctx.AbortWithStatusJSON(http.StatusOK, updatedUser.ResponseProfile(following))
}
