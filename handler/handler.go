package handler

import (
	"github.com/nathanbizkit/article-management-go/auth"
	"github.com/nathanbizkit/article-management-go/env"
	"github.com/nathanbizkit/article-management-go/store"
	"github.com/rs/zerolog"
)

// Handler definition
type Handler struct {
	logger  *zerolog.Logger
	environ *env.ENV
	authen  *auth.Auth
	us      *store.UserStore
	as      *store.ArticleStore
}

// New returns a new handler with logger, env, auth and stores
func New(l *zerolog.Logger, environ *env.ENV, authen *auth.Auth, us *store.UserStore, as *store.ArticleStore) *Handler {
	return &Handler{logger: l, environ: environ, authen: authen, us: us, as: as}
}
