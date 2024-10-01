package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nathanbizkit/article-management/messages"
)

// GetTags returns all of tags
func (h *Handler) GetTags(ctx *gin.Context) {
	h.logger.Info().Interface("req", ctx.Request).Msg("get tags")

	tags, err := h.as.GetTags(ctx.Request.Context())
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to get tags")
		ctx.AbortWithStatusJSON(http.StatusInternalServerError,
			gin.H{"error": "internal server error"})
		return
	}

	tagNames := make([]string, 0, len(tags))
	for _, t := range tags {
		tagNames = append(tagNames, t.Name)
	}

	ctx.JSON(http.StatusOK, gin.H{"payload": messages.TagsResponse{Tags: tagNames}})
}
