package user_controller_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"fullcycle-auction_go/internal/infra/api/web/controller/user_controller"
	"fullcycle-auction_go/internal/internal_error"
	"fullcycle-auction_go/internal/usecase/user_usecase"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockUserUseCase struct {
	mock.Mock
}

func (m *mockUserUseCase) FindUserById(
	ctx context.Context, id string) (*user_usecase.UserOutputDTO, *internal_error.InternalError) {
	args := m.Called(ctx, id)

	var out *user_usecase.UserOutputDTO
	if v := args.Get(0); v != nil {
		out = v.(*user_usecase.UserOutputDTO)
	}

	var err *internal_error.InternalError
	if v := args.Get(1); v != nil {
		err = v.(*internal_error.InternalError)
	}

	return out, err
}

func init() {
	gin.SetMode(gin.TestMode)
}

func setupUserRouter(uc user_usecase.UserUseCaseInterface) *gin.Engine {
	r := gin.New()
	r.GET("/users/:userId", user_controller.NewUserController(uc).FindUserById)
	return r
}

func TestFindUserById_InvalidUUID_ReturnsBadRequest(t *testing.T) {
	t.Parallel()

	useCase := new(mockUserUseCase)
	router := setupUserRouter(useCase)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/users/not-a-uuid", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	// validação de UUID acontece no controller, antes do use case.
	useCase.AssertNotCalled(t, "FindUserById", mock.Anything, mock.Anything)
}

func TestFindUserById_Found_ReturnsOK(t *testing.T) {
	t.Parallel()

	id := uuid.NewString()
	useCase := new(mockUserUseCase)
	useCase.On("FindUserById", mock.Anything, id).
		Return(&user_usecase.UserOutputDTO{Id: id, Name: "Jane Doe"}, nil)
	router := setupUserRouter(useCase)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/users/"+id, nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var body user_usecase.UserOutputDTO
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.Equal(t, id, body.Id)
	require.Equal(t, "Jane Doe", body.Name)
	useCase.AssertExpectations(t)
}

func TestFindUserById_UseCaseNotFound_ReturnsNotFound(t *testing.T) {
	t.Parallel()

	id := uuid.NewString()
	useCase := new(mockUserUseCase)
	useCase.On("FindUserById", mock.Anything, id).
		Return(nil, internal_error.NewNotFoundError("user not found"))
	router := setupUserRouter(useCase)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/users/"+id, nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
	useCase.AssertExpectations(t)
}
