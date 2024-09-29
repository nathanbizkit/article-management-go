package store

import "database/sql"

// ArticleStore is a data access struct for articles
type ArticleStore struct {
	db *sql.DB
}

// NewArticleStore returns a new ArticleStore
func NewArticleStore(db *sql.DB) *ArticleStore {
	return &ArticleStore{db: db}
}
