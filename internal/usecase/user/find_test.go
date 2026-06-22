package user_test

import (
	"context"
	"testing"

	"fullcycle-auction_go/internal/entity/user"
	"fullcycle-auction_go/internal/internal_error"
	useruc "fullcycle-auction_go/internal/usecase/user"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockUserRepository struct {
	mock.Mock
}

func (m *mockUserRepository) FindUserByID(ctx context.Context, userID string) (*user.User, *internal_error.InternalError) {
	args := m.Called(ctx, userID)

	var u *user.User
	if value := args.Get(0); value != nil {
		u, _ = value.(*user.User)
	}

	var err *internal_error.InternalError
	if value := args.Get(1); value != nil {
		err, _ = value.(*internal_error.InternalError)
	}

	return u, err
}

func TestFindUserByID(t *testing.T) {
	t.Parallel()

	t.Run("returns DTO", func(t *testing.T) {
		t.Parallel()
		repository := new(mockUserRepository)
		repository.On("FindUserByID", mock.Anything, "user-123").Return(&user.User{
			ID:   "user-123",
			Name: "Jane Doe",
		}, nil)

		useCase := useruc.New(repository)
		result, err := useCase.FindUserByID(context.Background(), "user-123")

		require.Nil(t, err)
		require.NotNil(t, result)
		require.Equal(t, "user-123", result.ID)
		require.Equal(t, "Jane Doe", result.Name)
		repository.AssertExpectations(t)
	})

	t.Run("propagates repository error", func(t *testing.T) {
		t.Parallel()
		repository := new(mockUserRepository)
		expectedError := internal_error.NewNotFoundError("user not found")
		repository.On("FindUserByID", mock.Anything, "user-123").Return(nil, expectedError)

		useCase := useruc.New(repository)
		result, err := useCase.FindUserByID(context.Background(), "user-123")

		require.Nil(t, result)
		require.ErrorIs(t, err, expectedError)
		repository.AssertExpectations(t)
	})
}