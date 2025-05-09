package main

import (
	"net/http"
	"time"

	"github.com/9thDuck/chat_go.git/cmd/api/ws"
	"github.com/9thDuck/chat_go.git/internal/auth"
	cloudStorage "github.com/9thDuck/chat_go.git/internal/cloud_storage"
	"github.com/9thDuck/chat_go.git/internal/store"
	"github.com/9thDuck/chat_go.git/internal/store/cache"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"go.uber.org/zap"
)

type application struct {
	config        config
	store         store.Storage
	cache         cache.Storage
	logger        *zap.SugaredLogger
	authenticator auth.Authenticator
	socketHub     *ws.Hub
	cloud         *cloudStorage.CloudStorage
}

type ctxKey string

func (app *application) mount() http.Handler {
	handler := chi.NewRouter()

	handler.Use(middleware.Logger)
	handler.Use(middleware.RealIP)
	handler.Use(middleware.RequestID)
	handler.Use(middleware.Recoverer)
	handler.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, //
	}))

	handler.Route("/v1", func(r chi.Router) {
		r.Get("/", app.getHomeHandler)

		r.Route("/auth", func(r chi.Router) {
			r.Post("/signup", app.signupHandler)
			r.Post("/login", app.loginHandler)
			r.With(app.ValidateTokenMiddleware()).Delete("/logout", app.logoutHandler)
		})

		r.Route("/users", func(r chi.Router) {
			r.Use(app.ValidateTokenMiddleware())
			r.With(app.encryptionIDMiddleware).Get("/", app.getAuthenticatedUserHandler)

			r.Route("/search", func(r chi.Router) {
				r.With(app.paginationMiddleware).Get("/", app.searchUsersByUsernamesAndGetIDsHandler)
				r.NotFound(func(w http.ResponseWriter, r *http.Request) {
					app.notFoundError(w, r, nil, "this api route is not valid")
				})
			})

			// User ID specific routes
			r.Route("/{userID}", func(r chi.Router) {
				r.Use(app.getUserIDParamMiddleware)
				r.Patch("/", app.userDetailsUpdateGuardMiddleware(app.updateUserByIDHandler))
			})
		})

		r.Route("/contacts", func(r chi.Router) {
			r.Use(app.ValidateTokenMiddleware())
			r.With(app.paginationMiddleware).Get("/", app.getContactsHandler)
			r.With(app.paginationMiddleware).Get("/search", app.searchContactsHandler)
			r.Route("/{contactID}", func(r chi.Router) {
				r.Use(app.getContactIDParamMiddleware)
				r.Delete("/", app.deleteContactHandler)
			})
			// requests
			r.Route("/requests", func(r chi.Router) {
				r.With(app.paginationMiddleware).Get("/", app.getContactRequestByIDHandler)
				r.Route("/{contactID}", func(r chi.Router) {
					r.Use(app.getContactIDParamMiddleware)
					r.Use(app.blockSelfContactRequestMiddleware)
					r.Post("/", app.createContactRequestHandler)
					r.Patch("/", app.updateContactRequestHandler)
					r.Delete("/", app.deleteContactRequestHandler)
				})
			})
		})

		r.Route("/messages", func(r chi.Router) {
			r.Use(app.ValidateTokenMiddleware())
			r.With(app.paginationMiddleware).Get("/", app.getMessagesHandler)
			r.Route("/{receiverID}", func(r chi.Router) {
				r.Use(app.getReceiverIDParamMiddleware)
				r.With(app.preMessageCreationMiddleware).Post("/", app.createMessageHandler)
			})
		})

		r.Route("/ws", func(r chi.Router) {
			r.Use(app.ValidateTokenMiddleware())
			r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				userID := getUserFromCtx(r).ID
				ws.Serve(w, r, app.socketHub, userID)
			})
		})

		r.Route("/cloud", func(r chi.Router) {
			r.Use(app.ValidateTokenMiddleware())
			r.Route("/presignedurl", func(r chi.Router) {
				r.Get("/{objectKey}", app.getPresignedS3URLHandler)
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
	app.socketHub = ws.NewHub()
	go app.socketHub.Run()
	app.logger.Infow("Server listening", "port", app.config.addr)

	return srv.ListenAndServe()
}
