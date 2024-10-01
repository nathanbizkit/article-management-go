package message

import "time"

/* Request message */

// CreateArticleRequest definition
type CreateArticleRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Body        string   `json:"body"`
	Tags        []string `json:"tags"`
}

// UpdateArticleRequest definition
type UpdateArticleRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Body        string `json:"body"`
}

// CreateCommentRequest definition
type CreateCommentRequest struct {
	Body string `json:"body"`
}

/* Response message */

// ArticleResponse definition
type ArticleResponse struct {
	ID             uint            `json:"id"`
	Title          string          `json:"title"`
	Description    string          `json:"description"`
	Body           string          `json:"body"`
	Tags           []string        `json:"tags"`
	Favorited      bool            `json:"favorited"`
	FavoritesCount int64           `json:"favorites_count"`
	Author         ProfileResponse `json:"author"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      *time.Time      `json:"updated_at"`
}

// ArticlesResponse definition
type ArticlesResponse struct {
	Articles      []ArticleResponse `json:"articles"`
	ArticlesCount int64             `json:"articles_count"`
}

// TagsResponse definition
type TagsResponse struct {
	Tags []string `json:"tags"`
}

// CommentResponse definition
type CommentResponse struct {
	ID        uint              `json:"id"`
	Body      string            `json:"body"`
	Author    []ProfileResponse `json:"author"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt *time.Time        `json:"updated_at"`
}

// CommentsResponse definition
type CommentsResponse struct {
	Comments []CommentResponse `json:"comments"`
}
