package handler

import (
	"github.com/nathanbizkit/article-management/auth"
	"github.com/nathanbizkit/article-management/env"
	"github.com/nathanbizkit/article-management/store"
	"github.com/rs/zerolog"
)

// Handler definition
type Handler struct {
	logger *zerolog.Logger
	env    *env.ENV
	auth   *auth.Auth
	us     *store.UserStore
	as     *store.ArticleStore
}

// New returns a new handler with logger, env, auth and database
func New(l *zerolog.Logger, e *env.ENV, auth *auth.Auth, us *store.UserStore, as *store.ArticleStore) *Handler {
	return &Handler{logger: l, env: e, auth: auth, us: us, as: as}
}
