package store

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"
	"github.com/nathanbizkit/article-management/db"
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
	var a model.Article
	var u model.User

	queryString := `SELECT 
		a.id, a.title, a.description, a.body, a.user_id, a.favorites_count, a.created_at, a.updated_at, 
		u.id, u.username, u.email, u.password, u.name, u.bio, u.image, u.created_at, u.updated_at 
		FROM article_management.articles a 
		INNER JOIN article_management.users u ON u.id = a.user_id 
		WHERE a.id = $1`
	err := s.db.QueryRowContext(ctx, queryString, id).
		Scan(
			&a.ID,
			&a.Title,
			&a.Description,
			&a.Body,
			&a.UserID,
			&a.FavoritesCount,
			&a.CreatedAt,
			&a.UpdatedAt,

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
		return nil, err
	}

	a.Author = u

	tags, err := getArticleTags(s.db, ctx, &a)
	if err != nil {
		return nil, err
	}

	a.Tags = tags

	return &a, nil
}

// Create creates an article and returns the newly created article
func (s *ArticleStore) Create(ctx context.Context, m *model.Article) (*model.Article, error) {
	var a model.Article

	err := db.RunInTx(s.db, func(tx *sql.Tx) error {
		queryString := `INSERT INTO article_management.articles (title, description, body, user_id) 
			VALUES ($1, $2, $3, $4) 
			RETURNING id, title, description, body, user_id, favorites_count, created_at, updated_at`
		err := tx.QueryRowContext(ctx, queryString, m.Title, m.Description, m.Body, m.UserID).
			Scan(
				&a.ID,
				&a.Title,
				&a.Description,
				&a.Body,
				&a.UserID,
				&a.FavoritesCount,
				&a.CreatedAt,
				&a.UpdatedAt,
			)
		if err != nil {
			return err
		}

		u, err := getArticleAuthor(s.db, ctx, &a)
		if err != nil {
			return err
		}

		a.Author = *u

		if len(m.Tags) > 0 {
			// create temporary tags table
			queryString = `CREATE TEMPORARY TABLE article_management.tags_staging 
				(LIKE article_management.tags INCLUDING ALL) ON COMMIT DROP`
			_, err := tx.ExecContext(ctx, queryString)
			if err != nil {
				return err
			}

			// copy into temporary table
			stmtTags, err := tx.PrepareContext(ctx,
				pq.CopyIn("article_management.tags_staging", "name"))
			if err != nil {
				return err
			}
			defer stmtTags.Close()

			for _, tag := range m.Tags {
				_, err := stmtTags.ExecContext(ctx, tag.Name)
				if err != nil {
					return err
				}
			}

			// insert into tags (on conflict do update)
			queryString = `INSERT INTO article_management.tags (name)
				SELECT name FROM article_management.tags_staging 
				ON CONFLICT (name) DO UPDATE SET updated_at = NOW() 
				RETURNING id, name, created_at, updated_at`
			rows, err := tx.QueryContext(ctx, queryString)
			if err != nil {
				return err
			}
			defer rows.Close()

			tags := make([]model.Tag, 0)
			for rows.Next() {
				var tag model.Tag

				err = rows.Scan(&tag.ID, &tag.Name, &tag.CreatedAt, &tag.UpdatedAt)
				if err != nil {
					return err
				}

				tags = append(tags, tag)
			}

			// insert into article_tags
			stmtArticleTags, err := tx.PrepareContext(ctx,
				pq.CopyIn("article_management.article_tags", "article_id", "tag_id"))
			if err != nil {
				return err
			}
			defer stmtArticleTags.Close()

			for _, tag := range m.Tags {
				_, err := stmtArticleTags.ExecContext(ctx, a.ID, tag.ID)
				if err != nil {
					return err
				}
			}

			a.Tags = tags
		}

		return nil
	})

	return &a, err
}

// Update updates an article (for title, description, body)
func (s *ArticleStore) Update(ctx context.Context, m *model.Article) (*model.Article, error) {
	var a model.Article

	err := db.RunInTx(s.db, func(tx *sql.Tx) error {
		queryString := `UPDATE article_management.articles 
			SET title = $1, description = $2, body = $3 
			WHERE id = $4 
			RETURNING id, title, description, body, user_id, favorites_count, created_at, updated_at`
		err := tx.QueryRowContext(ctx, queryString, m.Title, m.Description, m.Body, m.ID).
			Scan(
				&a.ID,
				&a.Title,
				&a.Description,
				&a.Body,
				&a.UserID,
				&a.FavoritesCount,
				&a.CreatedAt,
				&a.UpdatedAt,
			)
		if err != nil {
			return err
		}

		u, err := getArticleAuthor(s.db, ctx, &a)
		if err != nil {
			return err
		}

		a.Author = *u

		tags, err := getArticleTags(s.db, ctx, &a)
		if err != nil {
			return err
		}

		a.Tags = tags

		return nil
	})

	return &a, err
}

// GetArticles gets global articles
func (s *ArticleStore) GetArticles(ctx context.Context, tag, username string, favoritedBy *model.User, limit, offset int64) ([]model.Article, error) {
	var q bytes.Buffer
	q.WriteString(`SELECT 
		a.id, a.title, a.description, a.body, a.user_id, a.favorites_count, a.created_at, a.updated_at 
		u.id, u.username, u.email, u.password, u.name, u.bio, u.image, u.created_at, u.updated_at 
		FROM article_management.articles a 
		INNER JOIN article_management_users u ON u.id = a.user_id `)

	conds := make([]string, 0)
	condNumber := 1
	condItems := make([]interface{}, 0)

	if username != "" {
		conds = append(conds, fmt.Sprintf("u.username = $%d", condNumber))
		condItems = append(condItems, username)
		condNumber += 1
	}

	if tag != "" {
		q.WriteString(`INNER JOIN article_management.article_tags at ON at.article_id = a.id 
			INNER JOIN article_management.tags t ON t.id = at.tag_id `)

		conds = append(conds, fmt.Sprintf("t.name = $%d", condNumber))
		condItems = append(condItems, tag)
		condNumber += 1
	}

	if favoritedBy != nil {
		queryString := `SELECT article_id 
			FROM article_management.favorite_articles 
			WHERE user_id = $1 
			OFFSET $2 LIMIT $3`
		rows, err := s.db.QueryContext(ctx, queryString, favoritedBy.ID, offset, limit)
		if err != nil {
			return []model.Article{}, err
		}
		defer rows.Close()

		ids := make([]uint, 0)
		for rows.Next() {
			var id uint

			err = rows.Scan(&id)
			if err != nil {
				return []model.Article{}, err
			}

			ids = append(ids, id)
		}

		conds = append(conds, fmt.Sprintf("a.id in ($%d)", condNumber))
		condItems = append(condItems, pq.Array(ids))
		condNumber += 1
	}

	q.WriteString(strings.Join(conds, " AND "))

	rows, err := s.db.QueryContext(ctx, q.String(), condItems...)
	if err != nil {
		return []model.Article{}, err
	}
	defer rows.Close()

	as := make([]model.Article, 0)
	for rows.Next() {
		var a model.Article
		var u model.User

		err = rows.Scan(
			&a.ID,
			&a.Title,
			&a.Description,
			&a.Body,
			&a.UserID,
			&a.FavoritesCount,
			&a.CreatedAt,
			&a.UpdatedAt,

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
			return []model.Article{}, err
		}

		a.Author = u
		as = append(as, a)
	}

	articles, err := getArticlesTags(s.db, ctx, as)
	if err != nil {
		return []model.Article{}, err
	}

	return articles, nil
}

// GetFeedArticles gets following users' articles
func (s *ArticleStore) GetFeedArticles(ctx context.Context, userIDs []uint, limit, offset int64) ([]model.Article, error) {
	queryString := `SELECT 
		a.id, a.title, a.description, a.body, a.user_id, a.favorites_count, a.created_at, a.updated_at 
		u.id, u.username, u.email, u.password, u.name, u.bio, u.image, u.created_at, u.updated_at 
		FROM article_management.articles a 
		INNER JOIN article_management_users u ON u.id = a.user_id 
		WHERE a.user_id IN ($1) 
		OFFSET $2 LIMIT $3`
	rows, err := s.db.QueryContext(ctx, queryString, pq.Array(userIDs), offset, limit)
	if err != nil {
		return []model.Article{}, nil
	}

	as := make([]model.Article, 0)
	for rows.Next() {
		var a model.Article
		var u model.User

		err = rows.Scan(
			&a.ID,
			&a.Title,
			&a.Description,
			&a.Body,
			&a.UserID,
			&a.FavoritesCount,
			&a.CreatedAt,
			&a.UpdatedAt,

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
			return []model.Article{}, err
		}

		a.Author = u
		as = append(as, a)
	}

	articles, err := getArticlesTags(s.db, ctx, as)
	if err != nil {
		return []model.Article{}, err
	}

	return articles, nil
}

// Delete deletes an article
func (s *ArticleStore) Delete(ctx context.Context, m *model.Article) error {
	return db.RunInTx(s.db, func(tx *sql.Tx) error {
		queryString := `DELETE FROM article_management.articles WHERE id = $1`
		_, err := tx.ExecContext(ctx, queryString, m.ID)
		return err
	})
}

// IsFavorited checks whether the article is favorited by the user
func (s *ArticleStore) IsFavorited(ctx context.Context, a *model.Article, u *model.User) (bool, error) {
	if a == nil || u == nil {
		return false, nil
	}

	var count int

	queryString := `SELECT COUNT(article_id) 
		FROM article_management.favorite_articles 
		WHERE article_id = $1 AND user_id = $2`
	err := s.db.QueryRowContext(ctx, queryString, a.ID, u.ID).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// AddFavorite favorites an article
func (s *ArticleStore) AddFavorite(ctx context.Context, a *model.Article, u *model.User, updateFunc func(favoritesCount int64)) error {
	return db.RunInTx(s.db, func(tx *sql.Tx) error {
		queryString := `INSERT INTO article_management.favorite_articles (article_id, user_id) VALUES ($1, $2)`
		_, err := tx.ExecContext(ctx, queryString, a.ID, u.ID)
		if err != nil {
			return err
		}

		queryString = `UPDATE article_management.articles 
			SET favorites_count = favorites_count + $1 WHERE id = $2`
		_, err = tx.ExecContext(ctx, queryString, 1, a.ID)
		if err != nil {
			return err
		}

		updateFunc(a.FavoritesCount + 1)

		return nil
	})
}

// DeleteFavorite unfavorites an article
func (s *ArticleStore) DeleteFavorite(ctx context.Context, a *model.Article, u *model.User, updateFunc func(favoritesCount int64)) error {
	return db.RunInTx(s.db, func(tx *sql.Tx) error {
		queryString := `DELETE FROM article_management.favorite_articles WHERE article_id = $1 AND user_id = $2`
		_, err := tx.ExecContext(ctx, queryString, a.ID, u.ID)
		if err != nil {
			return err
		}

		queryString = `UPDATE article_management.articles 
			SET favorites_count = favorites_count - $1 WHERE id = $2`
		_, err = tx.ExecContext(ctx, queryString, 1, a.ID)
		if err != nil {
			return err
		}

		if a.FavoritesCount == 0 {
			updateFunc(0)
		} else {
			updateFunc(a.FavoritesCount - 1)
		}

		return nil
	})
}

// GetTags gets all tags
func (s *ArticleStore) GetTags(ctx context.Context) ([]model.Tag, error) {
	queryString := `SELECT id, name, created_at, updated_at FROM article_management.tags`
	rows, err := s.db.QueryContext(ctx, queryString)
	if err != nil {
		return []model.Tag{}, nil
	}
	defer rows.Close()

	tags := make([]model.Tag, 0)
	for rows.Next() {
		var tag model.Tag

		err = rows.Scan(&tag.ID, &tag.Name, &tag.CreatedAt, &tag.UpdatedAt)
		if err != nil {
			return []model.Tag{}, err
		}

		tags = append(tags, tag)
	}

	return tags, nil
}

// CreateComment creates a comment of the article
func (s *ArticleStore) CreateComment(ctx context.Context, m *model.Comment) (*model.Comment, error) {
	var c model.Comment

	err := db.RunInTx(s.db, func(tx *sql.Tx) error {
		queryString := `INSERT INTO article_management.comments (body, user_id, article_id) 
			VALUES ($1, $2, $3) 
			RETURNING id, body, user_id, article_id, created_at, updated_at`
		err := tx.QueryRowContext(ctx, queryString, m.Body, m.UserID, m.ArticleID).
			Scan(&c.ID, &c.Body, &c.UserID, &c.ArticleID, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return err
		}

		var u model.User

		queryString = `SELECT id, username, email, password, name, bio, image, created_at, updated_at 
			FROM article_management.users WHERE id = $1`
		err = tx.QueryRowContext(ctx, queryString, m.UserID).
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
			return err
		}

		c.Author = u

		return nil
	})

	return &c, err
}

// GetComments gets comments of the article
func (s *ArticleStore) GetComments(ctx context.Context, m *model.Article) ([]model.Comment, error) {
	queryString := `SELECT 
		c.id, c.body, c.user_id, c.article_id, c.created_at, c.updated_at, 
		u.id, u.username, u.email, u.password, u.name, u.bio, u.image, u.created_at, u.updated_at, 
		FROM article_management.comments c 
		INNER JOIN article_management.users u ON u.id = c.user_id 
		WHERE c.article_id = $1`
	rows, err := s.db.QueryContext(ctx, queryString, m.ID)
	if err != nil {
		return []model.Comment{}, nil
	}

	cs := make([]model.Comment, 0)
	for rows.Next() {
		var c model.Comment
		var u model.User

		err = rows.Scan(
			&c.ID,
			&c.Body,
			&c.UserID,
			&c.ArticleID,
			&c.CreatedAt,
			&c.UpdatedAt,

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
			return []model.Comment{}, nil
		}

		c.Author = u
		cs = append(cs, c)
	}

	return cs, nil
}

// GetCommentByID finds a comment from id
func (s *ArticleStore) GetCommentByID(ctx context.Context, id uint) (*model.Comment, error) {
	var c model.Comment
	var u model.User

	queryString := `SELECT 
		c.id, c.body, c.user_id, c.article_id, c.created_at, c.updated_at, 
		u.id, u.username, u.email, u.password, u.name, u.bio, u.image, u.created_at, u.updated_at, 
		FROM article_management.comments c 
		INNER JOIN article_management.users u ON u.id = c.user_id 
		WHERE c.id = $1`
	err := s.db.QueryRowContext(ctx, queryString, id).
		Scan(
			&c.ID,
			&c.Body,
			&c.UserID,
			&c.ArticleID,
			&c.CreatedAt,
			&c.UpdatedAt,

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
		return nil, err
	}

	c.Author = u
	return &c, nil
}

// DeleteComment deletes a comment
func (s *ArticleStore) DeleteComment(ctx context.Context, m *model.Comment) error {
	return db.RunInTx(s.db, func(tx *sql.Tx) error {
		queryString := `DELETE FROM article_management.comments WHERE id = $1`
		_, err := tx.ExecContext(ctx, queryString, m.ID)
		return err
	})
}

func getArticleAuthor(db *sql.DB, ctx context.Context, a *model.Article) (*model.User, error) {
	var u model.User

	queryString := `SELECT 
		u.id, u.username, u.email, u.password, u.name, u.bio, u.image, u.created_at, u.updated_at 
		FROM article_management.users u 
		INNER JOIN article_management.articles a ON a.user_id = u.id 
		WHERE a.id = $1`
	err := db.QueryRowContext(ctx, queryString, a.ID).
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
		return nil, err
	}

	return &u, nil
}

func getArticleTags(db *sql.DB, ctx context.Context, a *model.Article) ([]model.Tag, error) {
	queryString := `SELECT t.id, t.name, t.created_at, t.updated_at 
		FROM article_management.tags t 
		INNER JOIN article_management.article_tags at ON at.tag_id = t.id 
		WHERE at.article_id = $1`
	rows, err := db.QueryContext(ctx, queryString, a.ID)
	if err != nil {
		return []model.Tag{}, err
	}
	defer rows.Close()

	tags := make([]model.Tag, 0)
	for rows.Next() {
		var tag model.Tag

		err = rows.Scan(&tag.ID, &tag.Name, &tag.CreatedAt, &tag.UpdatedAt)
		if err != nil {
			return []model.Tag{}, err
		}

		tags = append(tags, tag)
	}

	return tags, nil
}

func getArticlesTags(db *sql.DB, ctx context.Context, as []model.Article) ([]model.Article, error) {
	articles := make([]model.Article, 0, len(as))
	copy(articles, as)

	ids := make([]uint, 0, len(articles))
	for _, a := range articles {
		ids = append(ids, a.ID)
	}

	queryString := `SELECT a.id, t.id, t.name, t.created_at, t.updated_at 
		FROM (SELECT id FROM article_management.articles WHERE id IN ($1)) AS a 
		INNER JOIN article_management.article_tags at ON at.article_id = a.id 
		INNER JOIN article_management.tags t ON t.id = at.tag_id`
	rows, err := db.QueryContext(ctx, queryString, pq.Array(ids))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tagsMap := make(map[uint][]model.Tag)
	for rows.Next() {
		var articleID uint
		var tag model.Tag

		err = rows.Scan(&articleID, &tag.ID, &tag.Name, &tag.CreatedAt, &tag.UpdatedAt)
		if err != nil {
			return nil, err
		}

		if _, exists := tagsMap[articleID]; !exists {
			tagsMap[articleID] = make([]model.Tag, 0)
		}

		tagsMap[articleID] = append(tagsMap[articleID], tag)
	}

	for _, a := range articles {
		a.Tags = make([]model.Tag, 0)

		if tags, exists := tagsMap[a.ID]; exists {
			a.Tags = append(a.Tags, tags...)
		}
	}

	return articles, nil
}
