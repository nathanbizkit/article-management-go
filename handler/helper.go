package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nathanbizkit/article-management/model"
)

// GetCurrentUserOrAbort returns current auth user or abort
func (h *Handler) GetCurrentUserOrAbort(ctx *gin.Context) *model.User {
	id := h.auth.GetContextUserID(ctx)

	u, err := h.us.GetByID(ctx.Request.Context(), id)
	if err != nil {
		h.logger.Error().Err(err).Msg(fmt.Sprintf("current user (id=%d) not found", id))
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "current user not found"})
		return nil
	}

	return u
}

// GetParamIDOrAbort returns param value as uint id from url parameters or abort
func (h *Handler) GetParamAsIDOrAbort(ctx *gin.Context, key string) uint {
	value := ctx.Param(key)
	if len(value) == 0 {
		err := fmt.Errorf("param (%s) is empty", key)
		h.logger.Error().Err(err).Msg(err.Error())
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid %s id", key)})
		return 0
	}

	id, err := strconv.Atoi(value)
	if err != nil {
		h.logger.Error().Err(err).
			Msg(fmt.Sprintf("cannot convert %s (%s) into integer", key, value))
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid %s id", key)})
		return 0
	}

	return uint(id)
}