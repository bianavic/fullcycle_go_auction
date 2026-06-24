package bid_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"fullcycle-auction_go/internal/apperr"
	bidcontroller "fullcycle-auction_go/internal/infra/api/web/controller/bid"
	"fullcycle-auction_go/internal/usecase/bid"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockBidUseCase struct {
	mock.Mock
}

func (m *mockBidUseCase) CreateBid(
	ctx context.Context, input bid.InputDTO) *apperr.InternalError {
	args := m.Called(ctx, input)
	if v := args.Get(0); v != nil {
		return v.(*apperr.InternalError)
	}
	return nil
}

func (m *mockBidUseCase) FindWinningBidByAuctionID(
	ctx context.Context, auctionID string) (*bid.OutputDTO, *apperr.InternalError) {
	args := m.Called(ctx, auctionID)

	var out *bid.OutputDTO
	if v := args.Get(0); v != nil {
		out = v.(*bid.OutputDTO)
	}

	var err *apperr.InternalError
	if v := args.Get(1); v != nil {
		err = v.(*apperr.InternalError)
	}

	return out, err
}

func (m *mockBidUseCase) FindBidByAuctionID(
	ctx context.Context, auctionID string) ([]bid.OutputDTO, *apperr.InternalError) {
	args := m.Called(ctx, auctionID)

	var out []bid.OutputDTO
	if v := args.Get(0); v != nil {
		out = v.([]bid.OutputDTO)
	}

	var err *apperr.InternalError
	if v := args.Get(1); v != nil {
		err = v.(*apperr.InternalError)
	}

	return out, err
}

func init() {
	gin.SetMode(gin.TestMode)
}

func setupBidRouter(uc bid.UseCase) *gin.Engine {
	r := gin.New()
	c := bidcontroller.New(uc)
	r.POST("/bids", c.CreateBid)
	r.GET("/bids/:auctionId", c.FindBidByAuctionID)
	return r
}

func TestCreateBid(t *testing.T) {
	t.Parallel()

	t.Run("valid body returns created", func(t *testing.T) {
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
	})

	t.Run("malformed JSON returns bad request", func(t *testing.T) {
		t.Parallel()
		// BidInputDTO não tem binding tags; só JSON sintaticamente inválido falha o bind.
		useCase := new(mockBidUseCase)
		router := setupBidRouter(useCase)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/bids", strings.NewReader(`{`))
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
		useCase.AssertNotCalled(t, "CreateBid", mock.Anything, mock.Anything)
	})

	t.Run("type mismatch returns bad request", func(t *testing.T) {
		t.Parallel()
		useCase := new(mockBidUseCase)
		router := setupBidRouter(useCase)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/bids", strings.NewReader(`{"amount":"not-a-number"}`))
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
		useCase.AssertNotCalled(t, "CreateBid", mock.Anything, mock.Anything)
	})

	t.Run("use case error returns 500", func(t *testing.T) {
		t.Parallel()
		useCase := new(mockBidUseCase)
		useCase.On("CreateBid", mock.Anything, mock.Anything).
			Return(apperr.NewInternalServerError("boom"))
		router := setupBidRouter(useCase)

		body := `{"user_id":"` + uuid.NewString() + `","auction_id":"` + uuid.NewString() + `","amount":100}`
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/bids", strings.NewReader(body))
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusInternalServerError, w.Code)
		useCase.AssertExpectations(t)
	})
}

func TestFindBidByAuctionID(t *testing.T) {
	t.Parallel()

	t.Run("invalid UUID returns bad request", func(t *testing.T) {
		t.Parallel()
		useCase := new(mockBidUseCase)
		router := setupBidRouter(useCase)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/bids/not-a-uuid", nil)
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
		useCase.AssertNotCalled(t, "FindBidByAuctionID", mock.Anything, mock.Anything)
	})

	t.Run("found returns OK", func(t *testing.T) {
		t.Parallel()
		id := uuid.NewString()
		useCase := new(mockBidUseCase)
		useCase.On("FindBidByAuctionID", mock.Anything, id).
			Return([]bid.OutputDTO{{ID: uuid.NewString(), AuctionID: id, Amount: 100}}, nil)
		router := setupBidRouter(useCase)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/bids/"+id, nil)
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var body []bid.OutputDTO
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		require.Len(t, body, 1)
		require.Equal(t, id, body[0].AuctionID)
		useCase.AssertExpectations(t)
	})
}
