package handler

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestIntegration_ArticleHandler(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration tests.")
	}

	gin.SetMode("test")

	// h, lct := setUp(t)
	// log.Println(h, lct)

	t.Run("CreateArticle", func(t *testing.T) {})

	t.Run("GetArticle", func(t *testing.T) {})

	t.Run("GetArticles", func(t *testing.T) {})

	t.Run("GetFeedArticles", func(t *testing.T) {})

	t.Run("UpdateArticle", func(t *testing.T) {})

	t.Run("DeleteArticle", func(t *testing.T) {})

	t.Run("FavoriteArticle", func(t *testing.T) {})

	t.Run("UnfavoriteArticle", func(t *testing.T) {})
}
