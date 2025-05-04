package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/xsyro/goapi/internal/app/api/presenter"
	"github.com/xsyro/goapi/internal/app/api/requests"
	"github.com/xsyro/goapi/internal/app/service"
	"github.com/xsyro/goapi/internal/utils"
)

type todoHandler struct {
	tokenSvc  service.TokenServices
	todoSvc   service.TodoServices
	presenter presenter.Presenter
}

func NewTodoHander(tokenSvc service.TokenServices, todoSvc service.TodoServices,
	presenter presenter.Presenter) *todoHandler {
	return &todoHandler{
		tokenSvc:  tokenSvc,
		todoSvc:   todoSvc,
		presenter: presenter,
	}
}

func (h *todoHandler) Routes() chi.Router {
	r := chi.NewMux()

	// Protected Routes
	r.Group(func(r chi.Router) {
		r.Use(h.tokenSvc.Verifier())
		r.Use(h.tokenSvc.Validator)
		// Add protect routes/APIs here.
		r.Get("/", h.List())
		r.Get("/{id}", h.Get())
		r.Post("/", h.Create())
		r.Patch("/{id}", h.Patch())
		r.Delete("/{id}", h.Delete())
	})

	return r
}

func (h *todoHandler) List() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := utils.StartSpan(r.Context())
		defer span.End()

		todos, err := h.todoSvc.GetTodos(ctx)
		if err != nil {
			h.presenter.Error(w, r, err)
			return
		}

		h.presenter.JSON(w, r, todos)
	}
}

func (h *todoHandler) Create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := utils.StartSpan(r.Context())
		defer span.End()

		pr, problems, err := requests.DecodeValidRequest[requests.PostTodo](r)
		if err != nil {
			if problems != nil {
				h.presenter.Error(w, r, utils.ErrorInvalidUserInput(err, problems))
				return
			}
			h.presenter.Error(w, r, utils.ErrorBadRequest(err))
			return
		}

		t, err := h.todoSvc.CreateTodo(ctx, pr)
		if err != nil {
			h.presenter.Error(w, r, err)
			return
		}

		h.presenter.JSON(w, r, t)
	}
}

func (h *todoHandler) Patch() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := utils.StartSpan(r.Context())
		defer span.End()

		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			h.presenter.Error(w, r, utils.ErrorBadRequest(err))
			return
		}

		pr, problems, err := requests.DecodeValidRequest[requests.PatchTodo](r)
		if err != nil {
			if problems != nil {
				h.presenter.Error(w, r, utils.ErrorInvalidUserInput(err, problems))
				return
			}
			h.presenter.Error(w, r, utils.ErrorBadRequest(err))
			return
		}

		err = h.todoSvc.UpdateTodo(ctx, pr, id)
		if err != nil {
			h.presenter.Error(w, r, err)
			return
		}
	}
}

func (h *todoHandler) Get() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := utils.StartSpan(r.Context())
		defer span.End()

		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			h.presenter.Error(w, r, utils.ErrorBadRequest(err))
			return
		}

		todo, err := h.todoSvc.GetTodo(ctx, id)
		if err != nil {
			h.presenter.Error(w, r, err)
			return
		}

		h.presenter.JSON(w, r, todo)
	}
}

func (h *todoHandler) Delete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := utils.StartSpan(r.Context())
		defer span.End()

		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			h.presenter.Error(w, r, utils.ErrorBadRequest(err))
			return
		}

		err = h.todoSvc.DeleteTodo(ctx, id)
		if err != nil {
			h.presenter.Error(w, r, err)
			return
		}
	}
}
