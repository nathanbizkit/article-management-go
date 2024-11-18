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
	if value == "" {
		msg := fmt.Sprintf("invalid %s id", key)
		err := fmt.Errorf("param (%s) is empty", key)
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": msg})
		return 0
	}

	id, err := strconv.Atoi(value)
	if err != nil {
		h.logger.Error().Err(err).Msg(fmt.Sprintf("cannot convert %s (%s) into integer", key, value))
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid %s id", key)})
		return 0
	}

	return uint(id)
}

// GetPaginationQuery returns limit and offset queries from url
func (h *Handler) GetPaginationQuery(ctx *gin.Context, defaultLimit, defaultOffset int64) (limit, offset int64) {
	limit = defaultLimit
	offset = defaultOffset

	queryLimit := ctx.Query("limit")
	if queryLimit != "" {
		l, err := strconv.Atoi(queryLimit)
		if err == nil && l > 0 {
			limit = int64(l)
		}
	}

	queryOffset := ctx.Query("offset")
	if queryOffset != "" {
		o, err := strconv.Atoi(queryOffset)
		if err == nil && o > 0 {
			offset = int64(o)
		}
	}

	return
}
