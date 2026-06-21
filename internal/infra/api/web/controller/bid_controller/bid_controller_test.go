package bid_controller_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"fullcycle-auction_go/internal/infra/api/web/controller/bid_controller"
	"fullcycle-auction_go/internal/internal_error"
	"fullcycle-auction_go/internal/usecase/bid_usecase"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockBidUseCase struct {
	mock.Mock
}

func (m *mockBidUseCase) CreateBid(
	ctx context.Context, input bid_usecase.BidInputDTO) *internal_error.InternalError {
	args := m.Called(ctx, input)
	if v := args.Get(0); v != nil {
		ret, _ := v.(*internal_error.InternalError)
		return ret
	}
	return nil
}

func (m *mockBidUseCase) FindWinningBidByAuctionId(
	ctx context.Context, auctionId string) (*bid_usecase.BidOutputDTO, *internal_error.InternalError) {
	args := m.Called(ctx, auctionId)

	var out *bid_usecase.BidOutputDTO
	if v := args.Get(0); v != nil {
		out, _ = v.(*bid_usecase.BidOutputDTO)
	}

	var err *internal_error.InternalError
	if v := args.Get(1); v != nil {
		err, _ = v.(*internal_error.InternalError)
	}

	return out, err
}

func (m *mockBidUseCase) FindBidByAuctionId(
	ctx context.Context, auctionId string) ([]bid_usecase.BidOutputDTO, *internal_error.InternalError) {
	args := m.Called(ctx, auctionId)

	var out []bid_usecase.BidOutputDTO
	if v := args.Get(0); v != nil {
		out, _ = v.([]bid_usecase.BidOutputDTO)
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

func setupBidRouter(uc bid_usecase.BidUseCaseInterface) *gin.Engine {
	r := gin.New()
	c := bid_controller.NewBidController(uc)
	r.POST("/bids", c.CreateBid)
	r.GET("/bids/:auctionId", c.FindBidByAuctionId)
	return r
}

func TestCreateBid_ValidBody_ReturnsCreated(t *testing.T) {
	t.Parallel()

	useCase := new(mockBidUseCase)
	useCase.On("CreateBid", mock.Anything, mock.Anything).Return(nil)
	router := setupBidRouter(useCase)

	body := `{"user_id":"` + uuid.NewString() + `","auction_id":"` + uuid.NewString() + `","amount":100}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/bids", strings.NewReader(body))
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	useCase.AssertExpectations(t)
}

func TestCreateBid_MalformedJSON_ReturnsBadRequest(t *testing.T) {
	t.Parallel()

	// BidInputDTO não tem binding tags; só JSON sintaticamente inválido falha o bind.
	useCase := new(mockBidUseCase)
	router := setupBidRouter(useCase)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/bids", strings.NewReader(`{`))
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	useCase.AssertNotCalled(t, "CreateBid", mock.Anything, mock.Anything)
}

func TestCreateBid_TypeMismatch_ReturnsNotFound(t *testing.T) {
	t.Parallel()

	// QUIRK: erro de tipo no JSON (amount string) cai no ramo *json.UnmarshalTypeError
	// de validation.ValidateErr, que retorna NewNotFoundError -> 404 (não 400).
	// Teste trava o comportamento atual; distinto do JSON malformado (400).
	useCase := new(mockBidUseCase)
	router := setupBidRouter(useCase)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/bids", strings.NewReader(`{"amount":"not-a-number"}`))
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
	useCase.AssertNotCalled(t, "CreateBid", mock.Anything, mock.Anything)
}

func TestCreateBid_UseCaseError_ReturnsInternalServerError(t *testing.T) {
	t.Parallel()

	useCase := new(mockBidUseCase)
	useCase.On("CreateBid", mock.Anything, mock.Anything).
		Return(internal_error.NewInternalServerError("boom"))
	router := setupBidRouter(useCase)

	body := `{"user_id":"` + uuid.NewString() + `","auction_id":"` + uuid.NewString() + `","amount":100}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/bids", strings.NewReader(body))
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	useCase.AssertExpectations(t)
}

func TestFindBidByAuctionId_InvalidUUID_ReturnsBadRequest(t *testing.T) {
	t.Parallel()

	useCase := new(mockBidUseCase)
	router := setupBidRouter(useCase)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/bids/not-a-uuid", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	useCase.AssertNotCalled(t, "FindBidByAuctionId", mock.Anything, mock.Anything)
}

func TestFindBidByAuctionId_Found_ReturnsOK(t *testing.T) {
	t.Parallel()

	id := uuid.NewString()
	useCase := new(mockBidUseCase)
	useCase.On("FindBidByAuctionId", mock.Anything, id).
		Return([]bid_usecase.BidOutputDTO{{Id: uuid.NewString(), AuctionId: id, Amount: 100}}, nil)
	router := setupBidRouter(useCase)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/bids/"+id, nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var body []bid_usecase.BidOutputDTO
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.Len(t, body, 1)
	require.Equal(t, id, body[0].AuctionId)
	useCase.AssertExpectations(t)
}
