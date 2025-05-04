package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/xsyro/goapi/internal/app/api/presenter"
	"github.com/xsyro/goapi/internal/app/service"
	"github.com/xsyro/goapi/internal/utils"
)

type userHandler struct {
	tokenSvc  service.TokenServices
	userSvc   service.UserServices
	presenter presenter.Presenter
}

func NewUserHander(tokenSvc service.TokenServices, userSvc service.UserServices,
	presenter presenter.Presenter) *userHandler {
	return &userHandler{
		tokenSvc:  tokenSvc,
		userSvc:   userSvc,
		presenter: presenter,
	}
}

func (h *userHandler) Routes() chi.Router {
	r := chi.NewMux()

	// Protected Routes
	r.Group(func(r chi.Router) {
		r.Use(h.tokenSvc.Verifier())
		r.Use(h.tokenSvc.Validator)
		// Add protect routes/APIs here.
		r.Get("/", h.List())
		r.Get("/{id}", h.Get())
		r.Delete("/{id}", h.Delete())
	})

	return r
}

func (h *userHandler) List() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := utils.StartSpan(r.Context())
		defer span.End()

		users, err := h.userSvc.GetUsers(ctx)
		if err != nil {
			h.presenter.Error(w, r, err)
			return
		}

		h.presenter.JSON(w, r, users)
	}
}

func (h *userHandler) Get() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := utils.StartSpan(r.Context())
		defer span.End()

		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			h.presenter.Error(w, r, utils.ErrorBadRequest(err))
			return
		}

		user, err := h.userSvc.GetUser(ctx, id)
		if err != nil {
			h.presenter.Error(w, r, err)
			return
		}

		h.presenter.JSON(w, r, user)
	}
}

func (h *userHandler) Delete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := utils.StartSpan(r.Context())
		defer span.End()

		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			h.presenter.Error(w, r, utils.ErrorBadRequest(err))
			return
		}

		err = h.userSvc.DeleteUser(ctx, id)
		if err != nil {
			h.presenter.Error(w, r, err)
			return
		}
	}
}
