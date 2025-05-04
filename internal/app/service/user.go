package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/xsyro/goapi/internal/app/api/requests"
	"github.com/xsyro/goapi/internal/app/repo"
	model "github.com/xsyro/goapi/internal/app/repo/sqlc"
	"github.com/xsyro/goapi/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	repo repo.Repository
}

func NewUserService(repo repo.Repository) *userService {
	return &userService{repo: repo}
}

func (s *userService) SignUp(ctx context.Context, sr requests.Signup) (model.Account, error) {
	ctx, span := utils.StartSpan(ctx)
	defer span.End()

	// Start a transaction
	tx, err := s.repo.Begin(ctx)
	if err != nil {
		return model.Account{}, repo.WrapDbError(err)
	}
	defer tx.Rollback(ctx)

	qtx := s.repo.WithTx(tx)
	account, err := qtx.CreateAccount(ctx, sr.Email)
	if err != nil {
		return model.Account{}, repo.WrapDbError(err)
	}

	password, err := bcrypt.GenerateFromPassword([]byte(sr.Password),
		bcrypt.DefaultCost)
	if err != nil {
		return model.Account{}, err
	}

	_, err = qtx.CreateUser(ctx, model.CreateUserParams{
		Name:      sr.Name,
		Email:     sr.Email,
		Password:  string(password),
		Role:      model.RoleAdmin,
		AccountID: account.ID,
	})
	if err != nil {
		return model.Account{}, repo.WrapDbError(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return model.Account{}, repo.WrapDbError(err)
	}

	return s.repo.GetAccountByID(ctx, account.ID)
}

func (s *userService) Authenticate(ctx context.Context, lr requests.Login) (model.User, error) {
	ctx, span := utils.StartSpan(ctx)
	defer span.End()

	user, err := s.repo.GetUserByEmail(ctx, lr.Email)
	if err != nil {
		return model.User{}, repo.WrapDbError(err)
	} else if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(lr.Password)) != nil {
		return model.User{}, utils.ErrInvalidCredentials
	}

	return user, nil
}

func (s *userService) GetUsers(ctx context.Context) ([]model.User, error) {
	ctx, span := utils.StartSpan(ctx)
	defer span.End()

	users, err := s.repo.GetUsers(ctx)
	if err != nil {
		return []model.User{}, repo.WrapDbError(err)
	}
	return users, nil
}

func (s *userService) GetUser(ctx context.Context, id uuid.UUID) (model.User, error) {
	ctx, span := utils.StartSpan(ctx)
	defer span.End()

	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return model.User{}, repo.WrapDbError(err)
	}
	return user, nil

}

func (s *userService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	ctx, span := utils.StartSpan(ctx)
	defer span.End()

	err := s.repo.DeleteUserByID(ctx, id)
	if err != nil {
		return repo.WrapDbError(err)
	}
	return nil

}
