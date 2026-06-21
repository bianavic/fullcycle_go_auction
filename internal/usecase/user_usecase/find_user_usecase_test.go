package user_usecase_test

import (
	"context"
	"fullcycle-auction_go/internal/entity/user_entity"
	"fullcycle-auction_go/internal/internal_error"
	"fullcycle-auction_go/internal/usecase/user_usecase"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockUserRepository struct {
	mock.Mock
}

func (m *mockUserRepository) FindUserById(ctx context.Context, userId string) (*user_entity.User, *internal_error.InternalError) {
	args := m.Called(ctx, userId)

	var user *user_entity.User
	if value := args.Get(0); value != nil {
		user = value.(*user_entity.User)
	}

	var err *internal_error.InternalError
	if value := args.Get(1); value != nil {
		err = value.(*internal_error.InternalError)
	}

	return user, err
}

func TestFindUserById_ShouldReturnUserOutputDTO_WhenRepositoryReturnsUser(t *testing.T) {
	t.Parallel()

	repository := new(mockUserRepository)
	repository.On("FindUserById", mock.Anything, "user-123").Return(&user_entity.User{
		Id:   "user-123",
		Name: "Jane Doe",
	}, nil)

	useCase := user_usecase.NewUserUseCase(repository)
	result, err := useCase.FindUserById(context.Background(), "user-123")

	require.Nil(t, err)
	require.NotNil(t, result)
	require.Equal(t, "user-123", result.Id)
	require.Equal(t, "Jane Doe", result.Name)

	repository.AssertExpectations(t)
}

func TestFindUserById_ShouldReturnError_WhenRepositoryFails(t *testing.T) {
	t.Parallel()

	repository := new(mockUserRepository)

	expectedError := internal_error.NewNotFoundError("user not found")

	repository.On("FindUserById", mock.Anything, "user-123").Return(nil, expectedError)

	useCase := user_usecase.NewUserUseCase(repository)

	result, err := useCase.FindUserById(context.Background(), "user-123")

	require.Nil(t, result)
	require.ErrorIs(t, err, expectedError)
	repository.AssertExpectations(t)
}
