package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nathanbizkit/article-management/message"
)

// GetTags returns all of tags
func (h *Handler) GetTags(ctx *gin.Context) {
	h.logger.Info().Msg("get tags")

	_ = h.GetCurrentUserOrAbort(ctx)

	tags, err := h.as.GetTags(ctx.Request.Context())
	if err != nil {
		msg := "failed to get tags"
		h.logger.Error().Err(err).Msg(msg)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}

	tagNames := make([]string, 0, len(tags))
	for _, t := range tags {
		tagNames = append(tagNames, t.Name)
	}

	ctx.JSON(http.StatusOK, message.TagsResponse{Tags: tagNames})
}
