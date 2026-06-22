package user_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"fullcycle-auction_go/internal/infra/api/web/controller/user"
	"fullcycle-auction_go/internal/internal_error"
	useruc "fullcycle-auction_go/internal/usecase/user"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockUserUseCase struct {
	mock.Mock
}

func (m *mockUserUseCase) FindUserByID(
	ctx context.Context, id string) (*useruc.UserOutputDTO, *internal_error.InternalError) {
	args := m.Called(ctx, id)

	var out *useruc.UserOutputDTO
	if v := args.Get(0); v != nil {
		out, _ = v.(*useruc.UserOutputDTO)
	}

	var err *internal_error.InternalError
	if v := args.Get(1); v != nil {
		err, _ = v.(*internal_error.InternalError)
	}

	return out, err
}

func init() {
	gin.SetMode(gin.TestMode)
}

func setupUserRouter(uc useruc.UseCase) *gin.Engine {
	r := gin.New()
	r.GET("/users/:userId", user.New(uc).FindUserByID)
	return r
}

func TestFindUserByID(t *testing.T) {
	t.Parallel()

	t.Run("invalid UUID returns bad request", func(t *testing.T) {
		t.Parallel()
		useCase := new(mockUserUseCase)
		router := setupUserRouter(useCase)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/users/not-a-uuid", nil)
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
		// validação de UUID acontece no controller, antes do use case.
		useCase.AssertNotCalled(t, "FindUserByID", mock.Anything, mock.Anything)
	})

	t.Run("found returns OK", func(t *testing.T) {
		t.Parallel()
		id := uuid.NewString()
		useCase := new(mockUserUseCase)
		useCase.On("FindUserByID", mock.Anything, id).
			Return(&useruc.UserOutputDTO{ID: id, Name: "Jane Doe"}, nil)
		router := setupUserRouter(useCase)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/users/"+id, nil)
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var body useruc.UserOutputDTO
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		require.Equal(t, id, body.ID)
		require.Equal(t, "Jane Doe", body.Name)
		useCase.AssertExpectations(t)
	})

	t.Run("use case not found returns 404", func(t *testing.T) {
		t.Parallel()
		id := uuid.NewString()
		useCase := new(mockUserUseCase)
		useCase.On("FindUserByID", mock.Anything, id).
			Return(nil, internal_error.NewNotFoundError("user not found"))
		router := setupUserRouter(useCase)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/users/"+id, nil)
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusNotFound, w.Code)
		useCase.AssertExpectations(t)
	})
}