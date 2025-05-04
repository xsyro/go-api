package handler

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"
	"github.com/rs/cors"
	"github.com/xsyro/goapi/config"
	"github.com/xsyro/goapi/docs"
	"github.com/xsyro/goapi/internal/app/api/auth"
	"github.com/xsyro/goapi/internal/app/api/mdlwr"
	"github.com/xsyro/goapi/internal/app/api/presenter"
	"github.com/xsyro/goapi/internal/app/repo"
	"github.com/xsyro/goapi/internal/app/service"
)

type Handler struct {
	*chi.Mux

	logger slog.Logger
	repo   repo.Repository
	redis  *redis.Client
}

func NewHandler(logger slog.Logger, repo repo.Repository, redis *redis.Client, cfg *config.Conf) *Handler {

	h := &Handler{
		Mux:    chi.NewMux(),
		logger: logger,
		repo:   repo,
		redis:  redis,
	}

	presenter := presenter.NewPresenters(logger)
	tokenAuth := auth.New(auth.HS256, []byte(cfg.AuthConfig.AccessSecret), nil, auth.AcceptableSkew)
	tokenSvc := service.NewTokenService(redis, cfg, tokenAuth, presenter)
	userSvc := service.NewUserService(repo)
	todoSvc := service.NewTodoService(repo)
	authhdlr := NewAuthHander(tokenSvc, userSvc, presenter)
	userhdlr := NewUserHander(tokenSvc, userSvc, presenter)
	todohdlr := NewTodoHander(tokenSvc, todoSvc, presenter)

	h.Use(middleware.RealIP)
	h.Use(mdlwr.RequestId)
	h.Use(mdlwr.Tracer)
	h.Use(mdlwr.Logger(logger))
	h.Use(cors.Default().Handler)
	h.Use(mdlwr.Recover(logger))

	// Serve the Swagger UI & openAPI spec
	h.Handle("/swagger/*", http.FileServer(http.FS(docs.SwaggerUI)))
	h.Get("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		w.Write(docs.OpenAPIYaml)
	})

	// Serve the APIs
	h.Mount("/", authhdlr.Routes())
	h.Mount("/users", userhdlr.Routes())
	h.Mount("/todos", todohdlr.Routes())
	return h
}
