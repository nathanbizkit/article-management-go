package store

import (
	"context"
	"database/sql"

	"github.com/nathanbizkit/article-management/model"
)

// ArticleStore is a data access struct for articles
type ArticleStore struct {
	db *sql.DB
}

// NewArticleStore returns a new ArticleStore
func NewArticleStore(db *sql.DB) *ArticleStore {
	return &ArticleStore{db: db}
}

// GetByID find an article by id
func (s *ArticleStore) GetByID(ctx context.Context, id uint) (*model.Article, error) {
	// TODO
	return nil, nil
}

// Create creates an article and returns the newly created article
func (s *ArticleStore) Create(ctx context.Context, m *model.Article) (*model.Article, error) {
	// TODO
	return nil, nil
}

// Update updates an article
func (s *ArticleStore) Update(ctx context.Context, m *model.Article, updateFunc func(a *model.Article)) error {
	// TODO
	return nil
}

// GetArticles gets global articles
func (s *ArticleStore) GetArticles(
	ctx context.Context, tagName, username string, favoritedBy *model.User, limit, offset int64) ([]model.Article, error) {

	// TODO
	return []model.Article{}, nil
}

// GetFeedArticles gets following users' articles
func (s *ArticleStore) GetFeedArticles(userIDs []uint, limit, offset int64) ([]model.Article, error) {
	// TODO
	return []model.Article{}, nil
}

// Delete deletes an article
func (s *ArticleStore) Delete(m *model.Article) error {
	// TODO
	return nil
}

// IsFavorited checks whether the article is favorited by the user
func (s *ArticleStore) IsFavorited(a *model.Article, u *model.User) (bool, error) {
	// TODO
	return false, nil
}

// AddFavorite favorites an article
func (s *ArticleStore) AddFavorite(a *model.Article, u *model.User, updateFunc func(u *model.User)) error {
	// TODO
	return nil
}

// DeleteFavorite unfavorites an article
func (s *ArticleStore) DeleteFavorite(a *model.Article, u *model.User, updateFunc func(u *model.User)) error {
	// TODO
	return nil
}

// GetTags gets all tags
func (s *ArticleStore) GetTags() ([]model.Tag, error) {
	// TODO
	return []model.Tag{}, nil
}

// CreateComment creates a comment of the article
func (s *ArticleStore) CreateComment(m *model.Comment) (*model.Comment, error) {
	// TODO
	return nil, nil
}

// GetComments gets comments of the article
func (s *ArticleStore) GetComments(m *model.Article) ([]model.Comment, error) {
	// TODO
	return []model.Comment{}, nil
}

// GetCommentByID finds a comment from id
func (s *ArticleStore) GetCommentByID(id uint) (*model.Comment, error) {
	// TODO
	return nil, nil
}

// DeleteComment deletes a comment
func (s *ArticleStore) DeleteComment(m *model.Comment) error {
	// TODO
	return nil
}
