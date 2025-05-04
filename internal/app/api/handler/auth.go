package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/xsyro/goapi/internal/app/api/auth"
	"github.com/xsyro/goapi/internal/app/api/presenter"
	"github.com/xsyro/goapi/internal/app/api/requests"
	"github.com/xsyro/goapi/internal/app/api/responses"
	model "github.com/xsyro/goapi/internal/app/repo/sqlc"
	"github.com/xsyro/goapi/internal/app/service"
	"github.com/xsyro/goapi/internal/utils"
)

type authHandler struct {
	tokenSvc  service.TokenServices
	userSvc   service.UserServices
	presenter presenter.Presenter
}

func NewAuthHander(tokenSvc service.TokenServices, userSvc service.UserServices,
	presenter presenter.Presenter) *authHandler {
	return &authHandler{
		tokenSvc:  tokenSvc,
		userSvc:   userSvc,
		presenter: presenter,
	}
}

func (h *authHandler) Routes() chi.Router {
	r := chi.NewMux()

	// Protected Routes
	r.Group(func(r chi.Router) {
		r.Use(h.tokenSvc.Verifier())
		r.Use(h.tokenSvc.Validator)
		// Add protect routes/APIs here.
		r.Post("/logout", h.logout)
	})

	// Public Routes
	r.Group(func(r chi.Router) {
		r.Post("/refresh", h.refresh)
		r.Post("/signup", h.signup)
		r.Post("/login", h.login)
	})

	return r
}

func (h *authHandler) signup(w http.ResponseWriter, r *http.Request) {
	ctx, span := utils.StartSpan(r.Context())
	defer span.End()

	sr, problems, err := requests.DecodeValidRequest[requests.Signup](r)
	if err != nil {
		if problems != nil {
			h.presenter.Error(w, r, utils.ErrorInvalidUserInput(err, problems))
			return
		}
		h.presenter.Error(w, r, utils.ErrorBadRequest(err))
		return
	}

	a, err := h.userSvc.SignUp(ctx, sr)
	if err != nil {
		if errors.Is(err, utils.ErrDuplicate) {
			h.presenter.Error(w, r, utils.ErrorDuplicate(err))
			return
		}
		h.presenter.Error(w, r, utils.ErrorInternal(err))
		return
	}

	h.presenter.JSON(w, r, a)
}

func (h *authHandler) login(w http.ResponseWriter, r *http.Request) {
	ctx, span := utils.StartSpan(r.Context())
	defer span.End()

	lr, problems, err := requests.DecodeValidRequest[requests.Login](r)
	if err != nil {
		if problems != nil {
			h.presenter.Error(w, r, utils.ErrorInvalidUserInput(err, problems))
			return
		}
		h.presenter.Error(w, r, utils.ErrorBadRequest(err))
		return
	}

	user, err := h.userSvc.Authenticate(ctx, lr)
	switch {
	case errors.Is(err, utils.ErrNotFound):
		h.presenter.Error(w, r, utils.ErrorNotFound(err))
		return
	case errors.Is(err, utils.ErrInvalidCredentials):
		h.presenter.Error(w, r, utils.ErrorInvalidCredentials(err))
		return
	case err != nil:
		h.presenter.Error(w, r, utils.ErrorInternal(err))
		return
	default:
		accessToken, refreshToken, exp, err := h.tokenSvc.GenerateTokenPair(ctx, &user)
		if err != nil {
			h.presenter.Error(w, r, utils.ErrorInternal(err))
			return
		}

		res := responses.NewLoginResponse(accessToken, refreshToken, int(exp.Unix()))
		h.presenter.JSON(w, r, res)
		return
	}
}

func (h *authHandler) refresh(w http.ResponseWriter, r *http.Request) {
	ctx, span := utils.StartSpan(r.Context())
	defer span.End()

	rr, problems, err := requests.DecodeValidRequest[requests.Refresh](r)
	if err != nil {
		if problems != nil {
			h.presenter.Error(w, r, utils.ErrorInvalidUserInput(err, problems))
			return
		}
		h.presenter.Error(w, r, utils.ErrorBadRequest(err))
		return
	}

	cusClaims, _ := h.tokenSvc.ParseToken(ctx, rr.Token)
	// Check for long inactivity even if the access token is not expired
	if err := h.tokenSvc.ValidateToken(r.Context(), cusClaims.Claims, true); err != nil {
		if errors.Is(err, auth.ErrNoTokenFound) {
			// Auto logout due to prolonged inactivity
			h.presenter.Error(w, r, utils.ErrorAuth(err))
			return
		}
		// Handle other token validation errors
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	u := &model.User{
		ID:    cusClaims.ID,
		Email: cusClaims.Username,
	}
	accessToken, refreshToken, exp, err := h.tokenSvc.GenerateTokenPair(ctx, u)
	if err != nil {
		h.presenter.Error(w, r, utils.ErrorInternal(err))
		return
	}

	res := responses.NewLoginResponse(accessToken, refreshToken, int(exp.Unix()))
	h.presenter.JSON(w, r, res)
}

func (h *authHandler) logout(w http.ResponseWriter, r *http.Request) {
	ctx, span := utils.StartSpan(r.Context())
	defer span.End()

	// remove the user token entry from Redis
	cusClaims, _ := h.tokenSvc.ParseToken(ctx, auth.TokenFromHeader(r))
	h.tokenSvc.RemoveToken(ctx, cusClaims.ID.String())
	h.presenter.JSON(w, r, "User logged out")
}
