package store

import (
	"context"
	"database/sql"

	"github.com/nathanbizkit/article-management/model"
)

// UserStore is a data access struct for users
type UserStore struct {
	db *sql.DB
}

// NewUserStore returns a new UserStore
func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{db: db}
}

// GetByID finds a user by id
func (s *UserStore) GetByID(ctx context.Context, id uint) (*model.User, error) {
	var m model.User
	queryString := `SELECT id, email, password, name, bio, image, created_at, updated_at 
		FROM article_management.users WHERE id = $1`
	err := s.db.QueryRowContext(ctx, queryString, id).
		Scan(
			&m.ID,
			&m.Email,
			&m.Password,
			&m.Name,
			&m.Bio,
			&m.Image,
			&m.CreatedAt,
			&m.UpdatedAt,
		)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// GetByEmail finds a user by email
func (s *UserStore) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var m model.User
	queryString := `SELECT id, email, password, name, bio, image, created_at, updated_at 
		FROM article_management.users WHERE email = $1`
	err := s.db.QueryRowContext(ctx, queryString, email).
		Scan(
			&m.ID,
			&m.Email,
			&m.Password,
			&m.Name,
			&m.Bio,
			&m.Image,
			&m.CreatedAt,
			&m.UpdatedAt,
		)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// GetByUsername finds a user by username
func (s *UserStore) GetByUsername(ctx context.Context, username uint) (*model.User, error) {
	var m model.User
	queryString := `SELECT id, email, password, name, bio, image, created_at, updated_at 
		FROM article_management.users WHERE username = $1`
	err := s.db.QueryRowContext(ctx, queryString, username).
		Scan(
			&m.ID,
			&m.Email,
			&m.Password,
			&m.Name,
			&m.Bio,
			&m.Image,
			&m.CreatedAt,
			&m.UpdatedAt,
		)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// Create creates a user and returns the newly created user
func (s *UserStore) Create(ctx context.Context, m *model.User) (*model.User, error) {
	// TODO
	return nil, nil
}

// Update updates a user
func (s *UserStore) Update(ctx context.Context, m *model.User, updateFunc func(u *model.User)) error {
	// TODO
	return nil
}

// IsFollowing returns wheter user A follows user B
func (s *UserStore) IsFollowing(ctx context.Context, a *model.User, b *model.User) (bool, error) {
	// TODO
	if a == nil || b == nil {
		return false, nil
	}
	return true, nil
}

// Follow creates a follow relationship from user A to user B
func (s *UserStore) Follow(ctx context.Context, a *model.User, b *model.User) error {
	// TODO
	return nil
}

// Unfollow deletes a follow relationship from user A to user B
func (s *UserStore) Unfollow(ctx context.Context, a *model.User, b *model.User) error {
	// TODO
	return nil
}

// GetFollowingUserIDs returns user ids that current user follows
func (s *UserStore) GetFollowingUserIDs(ctx context.Context, m *model.User) ([]uint, error) {
	// TODO
	return []uint{}, nil
}
