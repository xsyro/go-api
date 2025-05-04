package service_test

// import (
// 	"context"
// 	"testing"

// 	"github.com/xsyro/goapi/internal/app/api/requests"
// 	model "github.com/xsyro/goapi/internal/app/repo/sqlc"
// 	"github.com/xsyro/goapi/internal/app/service"
// 	mocks "github.com/xsyro/goapi/internal/mock"
// 	"github.com/google/uuid"
// 	"github.com/stretchr/testify/assert"
// 	"go.uber.org/mock/gomock"
// )

// func TestSignUp(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	mockRepo := mocks.NewMockRepository(ctrl)
// 	mockTx := mocks.NewMockTx(ctrl)
// 	userService := service.NewUserService(mockRepo)
// 	mockQueries := mocks.NewMockQueries(ctrl)

// 	ctx := context.Background()
// 	sr := requests.Signup{
// 		BasicAuth: requests.BasicAuth{
// 			Email:    "john@example.com",
// 			Password: "password123",
// 		},
// 		Name: "John Doe",
// 	}

// 	// Setting up the stub for WithTx method

// 	mockRepo.EXPECT().
// 		Begin(gomock.Any()). // Matches any context
// 		Return(mockTx, nil).
// 		Times(1)

// 	mockRepo.EXPECT().
// 		WithTx(mockTx).
// 		Return(mockQueries).
// 		Times(1) // Adjust Times as necessary based on your test scenario

// 	mockRepo.EXPECT().
// 		CreateAccount(gomock.Any(), sr.Email). // Matches any context
// 		Return(model.Account{ID: uuid.New(), Email: sr.Email}, nil).
// 		Times(1)

// 	mockRepo.EXPECT().
// 		CreateUser(gomock.Any(), gomock.Any()). // Matches any context and any second argument
// 		Return(model.User{ID: uuid.New(), Email: sr.Email, Name: sr.Name}, nil).
// 		Times(1)

// 	account, err := userService.SignUp(ctx, sr)
// 	assert.NoError(t, err)
// 	assert.Equal(t, sr.Email, account.Email)
// }
