package auction_controller_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"fullcycle-auction_go/internal/infra/api/web/controller/auction_controller"
	"fullcycle-auction_go/internal/internal_error"
	"fullcycle-auction_go/internal/usecase/auction_usecase"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockAuctionUseCase struct {
	mock.Mock
}

func (m *mockAuctionUseCase) CreateAuction(
	ctx context.Context, input auction_usecase.AuctionInputDTO) *internal_error.InternalError {
	args := m.Called(ctx, input)
	if v := args.Get(0); v != nil {
		ret, _ := v.(*internal_error.InternalError)
		return ret
	}
	return nil
}

func (m *mockAuctionUseCase) FindAuctionById(
	ctx context.Context, id string) (*auction_usecase.AuctionOutputDTO, *internal_error.InternalError) {
	args := m.Called(ctx, id)

	var out *auction_usecase.AuctionOutputDTO
	if v := args.Get(0); v != nil {
		out, _ = v.(*auction_usecase.AuctionOutputDTO)
	}

	var err *internal_error.InternalError
	if v := args.Get(1); v != nil {
		err, _ = v.(*internal_error.InternalError)
	}

	return out, err
}

func (m *mockAuctionUseCase) FindAuctions(
	ctx context.Context, status auction_usecase.AuctionStatus, category, productName string,
) ([]auction_usecase.AuctionOutputDTO, *internal_error.InternalError) {
	args := m.Called(ctx, status, category, productName)

	var out []auction_usecase.AuctionOutputDTO
	if v := args.Get(0); v != nil {
		out, _ = v.([]auction_usecase.AuctionOutputDTO)
	}

	var err *internal_error.InternalError
	if v := args.Get(1); v != nil {
		err, _ = v.(*internal_error.InternalError)
	}

	return out, err
}

func (m *mockAuctionUseCase) FindWinningBidByAuctionId(
	ctx context.Context, auctionId string) (*auction_usecase.WinningInfoOutputDTO, *internal_error.InternalError) {
	args := m.Called(ctx, auctionId)

	var out *auction_usecase.WinningInfoOutputDTO
	if v := args.Get(0); v != nil {
		out, _ = v.(*auction_usecase.WinningInfoOutputDTO)
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

func setupAuctionRouter(uc auction_usecase.AuctionUseCaseInterface) *gin.Engine {
	r := gin.New()
	c := auction_controller.NewAuctionController(uc)
	r.POST("/auctions", c.CreateAuction)
	r.GET("/auctions", c.FindAuctions)
	r.GET("/auctions/:auctionId", c.FindAuctionById)
	r.GET("/auctions/:auctionId/winner", c.FindWinningBidByAuctionId)
	return r
}

const validAuctionBody = `{"product_name":"Clock","category":"Decor","description":"A long enough description","condition":1}`

func TestCreateAuction_ValidBody_ReturnsCreated(t *testing.T) {
	t.Parallel()

	useCase := new(mockAuctionUseCase)
	useCase.On("CreateAuction", mock.Anything, mock.Anything).Return(nil)
	router := setupAuctionRouter(useCase)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auctions", strings.NewReader(validAuctionBody))
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	useCase.AssertExpectations(t)
}

func TestCreateAuction_BindingFailure_ReturnsBadRequest(t *testing.T) {
	t.Parallel()

	useCase := new(mockAuctionUseCase)
	router := setupAuctionRouter(useCase)

	// body vazio falha nos binding tags required/min do AuctionInputDTO.
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auctions", strings.NewReader(`{}`))
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	useCase.AssertNotCalled(t, "CreateAuction", mock.Anything, mock.Anything)
}

func TestCreateAuction_UseCaseInternalError_ReturnsInternalServerError(t *testing.T) {
	t.Parallel()

	useCase := new(mockAuctionUseCase)
	useCase.On("CreateAuction", mock.Anything, mock.Anything).
		Return(internal_error.NewInternalServerError("boom"))
	router := setupAuctionRouter(useCase)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auctions", strings.NewReader(validAuctionBody))
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	useCase.AssertExpectations(t)
}

func TestCreateAuction_UseCaseBadRequest_ReturnsBadRequest(t *testing.T) {
	t.Parallel()

	// cobre o ramo bad_request de rest_err.ConvertError.
	useCase := new(mockAuctionUseCase)
	useCase.On("CreateAuction", mock.Anything, mock.Anything).
		Return(internal_error.NewBadRequestError("invalid"))
	router := setupAuctionRouter(useCase)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auctions", strings.NewReader(validAuctionBody))
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	useCase.AssertExpectations(t)
}

func TestFindAuctionById_InvalidUUID_ReturnsBadRequest(t *testing.T) {
	t.Parallel()

	useCase := new(mockAuctionUseCase)
	router := setupAuctionRouter(useCase)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/auctions/not-a-uuid", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	useCase.AssertNotCalled(t, "FindAuctionById", mock.Anything, mock.Anything)
}

func TestFindAuctionById_Found_ReturnsOK(t *testing.T) {
	t.Parallel()

	id := uuid.NewString()
	useCase := new(mockAuctionUseCase)
	useCase.On("FindAuctionById", mock.Anything, id).
		Return(&auction_usecase.AuctionOutputDTO{Id: id, ProductName: "Clock"}, nil)
	router := setupAuctionRouter(useCase)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/auctions/"+id, nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var body auction_usecase.AuctionOutputDTO
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.Equal(t, id, body.Id)
	useCase.AssertExpectations(t)
}

func TestFindAuctionById_UseCaseNotFound_ReturnsNotFound(t *testing.T) {
	t.Parallel()

	id := uuid.NewString()
	useCase := new(mockAuctionUseCase)
	useCase.On("FindAuctionById", mock.Anything, id).
		Return(nil, internal_error.NewNotFoundError("missing"))
	router := setupAuctionRouter(useCase)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/auctions/"+id, nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
	useCase.AssertExpectations(t)
}

func TestFindAuctions_NonNumericStatus_ReturnsBadRequest(t *testing.T) {
	t.Parallel()

	useCase := new(mockAuctionUseCase)
	router := setupAuctionRouter(useCase)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/auctions?status=abc", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	useCase.AssertNotCalled(t, "FindAuctions", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestFindAuctions_EmptyStatus_ReturnsBadRequest(t *testing.T) {
	t.Parallel()

	// sem ?status= -> strconv.Atoi("") falha antes de chamar o use case.
	useCase := new(mockAuctionUseCase)
	router := setupAuctionRouter(useCase)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/auctions", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	useCase.AssertNotCalled(t, "FindAuctions", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestFindAuctions_ValidStatus_ReturnsOK(t *testing.T) {
	t.Parallel()

	useCase := new(mockAuctionUseCase)
	useCase.On("FindAuctions", mock.Anything, auction_usecase.AuctionStatus(0), "", "").
		Return([]auction_usecase.AuctionOutputDTO{{Id: uuid.NewString()}}, nil)
	router := setupAuctionRouter(useCase)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/auctions?status=0", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	useCase.AssertExpectations(t)
}

func TestFindWinningBid_InvalidUUID_ReturnsBadRequest(t *testing.T) {
	t.Parallel()

	useCase := new(mockAuctionUseCase)
	router := setupAuctionRouter(useCase)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/auctions/not-a-uuid/winner", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	useCase.AssertNotCalled(t, "FindWinningBidByAuctionId", mock.Anything, mock.Anything)
}

func TestFindWinningBid_FailOpenNilBid_ReturnsOKWithoutBidField(t *testing.T) {
	t.Parallel()

	// fail-open: use case retorna o auction com Bid nil; controller responde 200 e
	// o campo "bid" é omitido (json:"bid,omitempty").
	id := uuid.NewString()
	useCase := new(mockAuctionUseCase)
	useCase.On("FindWinningBidByAuctionId", mock.Anything, id).
		Return(&auction_usecase.WinningInfoOutputDTO{
			Auction: auction_usecase.AuctionOutputDTO{Id: id},
			Bid:     nil,
		}, nil)
	router := setupAuctionRouter(useCase)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/auctions/"+id+"/winner", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NotContains(t, w.Body.String(), `"bid"`)
	useCase.AssertExpectations(t)
}
