package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nanmu42/gzip"
	"github.com/nathanbizkit/article-management/auth"
	"github.com/nathanbizkit/article-management/db"
	"github.com/nathanbizkit/article-management/env"
	"github.com/nathanbizkit/article-management/handler"
	"github.com/nathanbizkit/article-management/middleware"
	"github.com/nathanbizkit/article-management/store"
	"github.com/rs/zerolog"
	_ "go.uber.org/automaxprocs"
)

func main() {
	w := zerolog.ConsoleWriter{Out: os.Stderr}
	l := zerolog.New(w).With().Timestamp().Caller().Logger()

	e, err := env.Parse()
	if err != nil {
		l.Fatal().Err(err).Msg("failed to load env")
	}

	l.Info().Msg("succeeded to load env")

	d, err := db.New(e)
	if err != nil {
		l.Fatal().Err(err).Msg("failed to connect to database")
	}

	l.Info().Str("name", "postgres").Str("database", e.DBName).
		Msg("succeeded to connect to database")

	if e.AppMode == "production" || e.AppMode == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	l.Info().Str("mode", e.AppMode).Msgf("setting app to %s mode", e.AppMode)
	l.Info().Str("mode", gin.Mode()).
		Msgf("gin router is running in %s mode", gin.Mode())

	router := gin.Default()
	router.Use(gzip.DefaultHandler().Gin)
	router.Use(middleware.CORS(e))

	auth := auth.New(e)

	us := store.NewUserStore(d)
	as := store.NewArticleStore(d)

	h := handler.New(&l, e, auth, us, as)
	handler.Route(router, auth, h)

	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", e.AppPort),
		Handler: router,
	}

	l.Info().Str("port", e.AppPort).Msg("starting server...")

	go func() {
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			l.Fatal().Err(err).Msg("failed to listen and serve")
		}
	}()

	<-ctx.Done()

	stop()
	l.Info().Msg("shutting down gracefully, press Ctrl+C again to force")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = d.Close()
	if err != nil {
		l.Fatal().Err(err).Msg("failed to close database connection")
	}

	err = srv.Shutdown(ctx)
	if err != nil {
		l.Fatal().Err(err).Msg("failed to shutdown http server")
	}

	l.Info().Msg("server exiting...")
}
