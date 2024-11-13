package handler

import (
	"log"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestIntegration_ProfileHandler(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration tests.")
	}

	gin.SetMode("test")

	h, lct := setUp(t)
	log.Println(h, lct)

	t.Run("ShowProfile", func(t *testing.T) {})

	t.Run("FollowUser", func(t *testing.T) {})

	t.Run("UnfollowUser", func(t *testing.T) {})
}
