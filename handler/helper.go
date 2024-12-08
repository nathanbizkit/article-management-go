package handler

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nathanbizkit/article-management-go/model"
)

// GetCurrentUserFromContext returns current auth user
func (h *Handler) GetCurrentUserFromContext(ctx *gin.Context) (*model.User, error) {
	return h.us.GetByID(ctx.Request.Context(), h.authen.GetContextUserID(ctx))
}

// GetIDFromParam returns param value as uint id from url parameters or abort
func (h *Handler) GetIDFromParam(ctx *gin.Context, key string) (uint, error) {
	value := ctx.Param(key)
	if value == "" {
		return 0, fmt.Errorf("param (%s) is empty", key)
	}

	id, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid %s id", key)
	}

	return uint(id), nil
}

// GetPaginationQuery returns limit and offset queries from url
func (h *Handler) GetPaginationQuery(ctx *gin.Context, defaultLimit, defaultOffset int64) (limit, offset int64) {
	limit = defaultLimit
	offset = defaultOffset

	queryLimit := ctx.Query("limit")
	if queryLimit != "" {
		l, err := strconv.Atoi(queryLimit)
		if err == nil && l != 0 {
			limit = int64(l)
		}
	}

	queryOffset := ctx.Query("offset")
	if queryOffset != "" {
		o, err := strconv.Atoi(queryOffset)
		if err == nil && o != 0 {
			offset = int64(o)
		}
	}

	return
}
