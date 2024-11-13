package handler

import (
	"log"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestIntegration_UserHandler(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration tests.")
	}

	gin.SetMode("test")

	h, lct := setUp(t)
	log.Println(h, lct)

	t.Run("Login", func(t *testing.T) {})

	t.Run("Register", func(t *testing.T) {})

	t.Run("RefreshToken", func(t *testing.T) {})

	t.Run("GetCurrentUser", func(t *testing.T) {})

	t.Run("UpdateCurrentUser", func(t *testing.T) {})
}
