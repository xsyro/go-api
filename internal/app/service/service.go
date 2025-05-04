package service

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/xsyro/goapi/internal/app/api/requests"
	model "github.com/xsyro/goapi/internal/app/repo/sqlc"
)

//go:generate mockgen -source=service.go -destination=../../mock/service.go -package=mocks
type UserServices interface {
	SignUp(ctx context.Context, sr requests.Signup) (model.Account, error)
	Authenticate(ctx context.Context, lr requests.Login) (model.User, error)
	GetUsers(ctx context.Context) ([]model.User, error)
	GetUser(ctx context.Context, id uuid.UUID) (model.User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
}

type TodoServices interface {
	GetTodos(ctx context.Context) ([]model.Todo, error)
	GetTodo(ctx context.Context, id uuid.UUID) (model.Todo, error)
	CreateTodo(ctx context.Context, tr requests.PostTodo) (model.Todo, error)
	UpdateTodo(ctx context.Context, tr requests.PatchTodo, id uuid.UUID) error
	DeleteTodo(ctx context.Context, id uuid.UUID) error
}

type TokenServices interface {
	GenerateTokenPair(ctx context.Context, user *model.User) (accessToken, refreshToken string, exp time.Time, err error)
	ParseToken(ctx context.Context, tokenString string) (claims *JwtCustomClaims, err error)
	ValidateToken(ctx context.Context, claims map[string]interface{}, isRefresh bool) error
	RemoveToken(ctx context.Context, userId string)
	ProlongToken(ctx context.Context, userId string)
	Verifier() func(http.Handler) http.Handler
	Validator(next http.Handler) http.Handler
}
