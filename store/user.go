package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/nathanbizkit/article-management/db"
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
	var u model.User

	queryString := `SELECT id, username, email, password, name, bio, image, created_at, updated_at 
		FROM article_management.users WHERE id = $1`
	err := s.db.QueryRowContext(ctx, queryString, id).
		Scan(
			&u.ID,
			&u.Username,
			&u.Email,
			&u.Password,
			&u.Name,
			&u.Bio,
			&u.Image,
			&u.CreatedAt,
			&u.UpdatedAt,
		)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = fmt.Errorf("failed to get user :%w", err)
		}
		return nil, err
	}

	return &u, nil
}

// GetByEmail finds a user by email
func (s *UserStore) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var u model.User

	queryString := `SELECT id, username, email, password, name, bio, image, created_at, updated_at 
		FROM article_management.users WHERE email = $1`
	err := s.db.QueryRowContext(ctx, queryString, email).
		Scan(
			&u.ID,
			&u.Username,
			&u.Email,
			&u.Password,
			&u.Name,
			&u.Bio,
			&u.Image,
			&u.CreatedAt,
			&u.UpdatedAt,
		)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = fmt.Errorf("failed to get user :%w", err)
		}
		return nil, err
	}

	return &u, nil
}

// GetByUsername finds a user by username
func (s *UserStore) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var u model.User

	queryString := `SELECT id, username, email, password, name, bio, image, created_at, updated_at 
		FROM article_management.users WHERE username = $1`
	err := s.db.QueryRowContext(ctx, queryString, username).
		Scan(
			&u.ID,
			&u.Username,
			&u.Email,
			&u.Password,
			&u.Name,
			&u.Bio,
			&u.Image,
			&u.CreatedAt,
			&u.UpdatedAt,
		)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = fmt.Errorf("failed to get user :%w", err)
		}
		return nil, err
	}

	return &u, nil
}

// Create creates a user and returns the newly created user
func (s *UserStore) Create(ctx context.Context, m *model.User) (*model.User, error) {
	var u model.User

	err := db.RunInTx(s.db, func(tx *sql.Tx) error {
		queryString := `INSERT INTO article_management.users (username, email, password, name, bio, image) 
			VALUES ($1, $2, $3, $4, $5, $6) 
			RETURNING id, username, email, password, name, bio, image, created_at, updated_at`
		err := tx.QueryRowContext(ctx, queryString, m.Username, m.Email, m.Password, m.Name, m.Bio, m.Image).
			Scan(
				&u.ID,
				&u.Username,
				&u.Email,
				&u.Password,
				&u.Name,
				&u.Bio,
				&u.Image,
				&u.CreatedAt,
				&u.UpdatedAt,
			)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				err = fmt.Errorf("failed to retrieve newly created user :%w", err)
			}
			return err
		}

		return nil
	})

	return &u, err
}

// Update updates a user (for username, email, password, name, bio, image)
func (s *UserStore) Update(ctx context.Context, m *model.User) (*model.User, error) {
	var u model.User

	err := db.RunInTx(s.db, func(tx *sql.Tx) error {
		queryString := `UPDATE article_management.users 
			SET username = $1, email = $2, password = $3, name = $4, bio = $5, image = $6 WHERE id = $7 
			RETURNING id, username, email, password, name, bio, image, created_at, updated_at`
		err := tx.QueryRowContext(ctx, queryString, m.Username, m.Email, m.Password, m.Name, m.Bio, m.Image, m.ID).
			Scan(
				&u.ID,
				&u.Username,
				&u.Email,
				&u.Password,
				&u.Name,
				&u.Bio,
				&u.Image,
				&u.CreatedAt,
				&u.UpdatedAt,
			)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				err = fmt.Errorf("failed to retrieve newly updated user :%w", err)
			}
			return err
		}

		return nil
	})

	return &u, err
}

// IsFollowing returns wheter user A follows user B
func (s *UserStore) IsFollowing(ctx context.Context, a *model.User, b *model.User) (bool, error) {
	if a == nil || b == nil {
		return false, nil
	}

	var count int

	queryString := `SELECT COUNT(to_user_id) FROM article_management.follows 
		WHERE from_user_id = $1 AND to_user_id = $2`
	err := s.db.QueryRowContext(ctx, queryString, a.ID, b.ID).Scan(&count)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false, err
	}

	return count > 0, nil
}

// Follow creates a follow relationship from user A to user B
func (s *UserStore) Follow(ctx context.Context, a *model.User, b *model.User) error {
	return db.RunInTx(s.db, func(tx *sql.Tx) error {
		queryString := `INSERT INTO article_management.follows (from_user_id, to_user_id) VALUES ($1, $2)`
		_, err := tx.ExecContext(ctx, queryString, a.ID, b.ID)
		return err
	})
}

// Unfollow deletes a follow relationship from user A to user B
func (s *UserStore) Unfollow(ctx context.Context, a *model.User, b *model.User) error {
	return db.RunInTx(s.db, func(tx *sql.Tx) error {
		queryString := `DELETE FROM article_management.follows WHERE from_user_id = $1 AND to_user_id = $2`
		_, err := tx.ExecContext(ctx, queryString, a.ID, b.ID)
		return err
	})
}

// GetFollowingUserIDs returns user ids that current user follows
func (s *UserStore) GetFollowingUserIDs(ctx context.Context, m *model.User) ([]uint, error) {
	queryString := `SELECT to_user_id FROM article_management.follows WHERE from_user_id = $1`
	rows, err := s.db.QueryContext(ctx, queryString, m.ID)
	if err != nil {
		return []uint{}, err
	}
	defer rows.Close()

	ids := make([]uint, 0)
	for rows.Next() {
		var id uint

		err = rows.Scan(&id)
		if err != nil {
			return []uint{}, err
		}

		ids = append(ids, id)
	}

	return ids, nil
}
