package main

import (
	"net/http"
	"time"

	"github.com/9thDuck/chat_go.git/internal/auth"
	"github.com/9thDuck/chat_go.git/internal/store"
	"github.com/9thDuck/chat_go.git/internal/store/cache"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type application struct {
	config        config
	store         store.Storage
	cache         cache.Storage
	logger        *zap.SugaredLogger
	authenticator auth.Authenticator
}

type ctxKey string

func (app *application) mount() http.Handler {
	handler := chi.NewRouter()

	handler.Use(middleware.Logger)
	handler.Use(middleware.RealIP)
	handler.Use(middleware.RequestID)
	handler.Use(middleware.Recoverer)

	handler.Route("/v1", func(r chi.Router) {
		r.Get("/", app.getHomeHandler)

		r.Route("/auth", func(r chi.Router) {
			r.Post("/signup", app.signupHandler)
			r.Post("/login", app.loginHandler)
			// r.Delete("/logout", app.logoutHandler)
		})

		r.Route("/users", func(r chi.Router) {
			r.Use(app.ValidateTokenMiddleware())
			r.Route("/{userID}", func(r chi.Router) {
				r.Use(app.getUserIDParamMiddleware)
				r.Get("/", app.getUserByIDHandler)
			})
		})
	})

	return handler
}

func (app *application) run(handler http.Handler) error {
	srv := &http.Server{
		Addr:         app.config.addr,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 30,
		IdleTimeout:  time.Minute,
		Handler:      handler,
	}
	app.logger.Infow("Server listening", "port", app.config.addr)

	return srv.ListenAndServe()
}
