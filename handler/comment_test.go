package handler

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestIntegration_CommentHandler(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration tests.")
	}

	gin.SetMode("test")

	// h, lct := setUp(t)
	// log.Println(h, lct)

	t.Run("CreateComment", func(t *testing.T) {})

	t.Run("GetComments", func(t *testing.T) {})

	t.Run("DeleteComment", func(t *testing.T) {})
}
