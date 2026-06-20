package auction_usecase_test

import (
	"context"
	"testing"
	"time"

	"fullcycle-auction_go/internal/entity/auction_entity"
	"fullcycle-auction_go/internal/entity/bid_entity"
	"fullcycle-auction_go/internal/internal_error"
	"fullcycle-auction_go/internal/usecase/auction_usecase"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// fakeAuctionRepo é um stub de auction_entity.AuctionRepositoryInterface.
type fakeAuctionRepo struct {
	created   []*auction_entity.Auction
	createErr *internal_error.InternalError

	byId    *auction_entity.Auction
	byIdErr *internal_error.InternalError

	list    []auction_entity.Auction
	listErr *internal_error.InternalError
}

func (f *fakeAuctionRepo) CreateAuction(ctx context.Context, a *auction_entity.Auction) *internal_error.InternalError {
	f.created = append(f.created, a)
	return f.createErr
}

func (f *fakeAuctionRepo) FindAuctions(ctx context.Context, status auction_entity.AuctionStatus, category, productName string) ([]auction_entity.Auction, *internal_error.InternalError) {
	return f.list, f.listErr
}

func (f *fakeAuctionRepo) FindAuctionById(ctx context.Context, id string) (*auction_entity.Auction, *internal_error.InternalError) {
	return f.byId, f.byIdErr
}

// fakeBidRepo é um stub de bid_entity.BidEntityRepository; só o caminho do lance
// vencedor é exercitado por estes testes.
type fakeBidRepo struct {
	winning    *bid_entity.Bid
	winningErr *internal_error.InternalError
}

func (f *fakeBidRepo) CreateBid(ctx context.Context, bids []bid_entity.Bid) *internal_error.InternalError {
	return nil
}

func (f *fakeBidRepo) FindBidByAuctionId(ctx context.Context, auctionId string) ([]bid_entity.Bid, *internal_error.InternalError) {
	return nil, nil
}

func (f *fakeBidRepo) FindWinningBidByAuctionId(ctx context.Context, auctionId string) (*bid_entity.Bid, *internal_error.InternalError) {
	return f.winning, f.winningErr
}

func validInput() auction_usecase.AuctionInputDTO {
	return auction_usecase.AuctionInputDTO{
		ProductName: "Clock",
		Category:    "Decor",
		Description: "A long enough description",
		Condition:   1,
	}
}

func TestCreateAuction_Success(t *testing.T) {
	t.Parallel()

	auctionRepo := &fakeAuctionRepo{}
	uc := auction_usecase.NewAuctionUseCase(auctionRepo, &fakeBidRepo{})

	err := uc.CreateAuction(context.Background(), validInput())
	require.Nil(t, err)
	require.Len(t, auctionRepo.created, 1)
	require.Equal(t, "Clock", auctionRepo.created[0].ProductName)
}

func TestCreateAuction_ValidationError(t *testing.T) {
	t.Parallel()

	auctionRepo := &fakeAuctionRepo{}
	uc := auction_usecase.NewAuctionUseCase(auctionRepo, &fakeBidRepo{})

	input := validInput()
	input.ProductName = "C" // muito curto -> falha na validação da entidade

	err := uc.CreateAuction(context.Background(), input)
	require.NotNil(t, err)
	require.Equal(t, "bad_request", err.Err)
	require.Empty(t, auctionRepo.created)
}

func TestCreateAuction_RepositoryError(t *testing.T) {
	t.Parallel()

	auctionRepo := &fakeAuctionRepo{createErr: internal_error.NewInternalServerError("unexpected error")}
	uc := auction_usecase.NewAuctionUseCase(auctionRepo, &fakeBidRepo{})

	err := uc.CreateAuction(context.Background(), validInput())
	require.NotNil(t, err)
	require.Equal(t, "internal_server_error", err.Err)
}

func TestFindAuctionById_MapsFields(t *testing.T) {
	t.Parallel()

	id := uuid.NewString()
	ts := time.Now()
	auctionRepo := &fakeAuctionRepo{byId: &auction_entity.Auction{
		Id: id, ProductName: "Clock", Category: "Decor",
		Description: "desc", Condition: auction_entity.New,
		Status: auction_entity.Active, Timestamp: ts,
	}}
	uc := auction_usecase.NewAuctionUseCase(auctionRepo, &fakeBidRepo{})

	out, err := uc.FindAuctionById(context.Background(), id)
	require.Nil(t, err)
	require.Equal(t, id, out.Id)
	require.Equal(t, "Clock", out.ProductName)
	require.Equal(t, auction_usecase.AuctionStatus(auction_entity.Active), out.Status)
	require.Equal(t, ts, out.Timestamp)
}

func TestFindAuctionById_RepositoryError(t *testing.T) {
	t.Parallel()

	auctionRepo := &fakeAuctionRepo{byIdErr: internal_error.NewNotFoundError("missing")}
	uc := auction_usecase.NewAuctionUseCase(auctionRepo, &fakeBidRepo{})

	out, err := uc.FindAuctionById(context.Background(), uuid.NewString())
	require.NotNil(t, err)
	require.Nil(t, out)
}

func TestFindAuctions_MapsFields(t *testing.T) {
	t.Parallel()

	auctionRepo := &fakeAuctionRepo{list: []auction_entity.Auction{
		{Id: uuid.NewString(), ProductName: "A", Status: auction_entity.Active},
		{Id: uuid.NewString(), ProductName: "B", Status: auction_entity.Completed},
	}}
	uc := auction_usecase.NewAuctionUseCase(auctionRepo, &fakeBidRepo{})

	out, err := uc.FindAuctions(context.Background(), 0, "", "")
	require.Nil(t, err)
	require.Len(t, out, 2)
}

func TestFindAuctions_RepositoryError(t *testing.T) {
	t.Parallel()

	auctionRepo := &fakeAuctionRepo{listErr: internal_error.NewInternalServerError("unexpected error")}
	uc := auction_usecase.NewAuctionUseCase(auctionRepo, &fakeBidRepo{})

	out, err := uc.FindAuctions(context.Background(), 0, "", "")
	require.NotNil(t, err)
	require.Nil(t, out)
}

func TestFindWinningBidByAuctionId_ReturnsBid(t *testing.T) {
	t.Parallel()

	id := uuid.NewString()
	auctionRepo := &fakeAuctionRepo{byId: &auction_entity.Auction{Id: id, Status: auction_entity.Completed}}
	bidRepo := &fakeBidRepo{winning: &bid_entity.Bid{Id: uuid.NewString(), AuctionId: id, Amount: 500}}
	uc := auction_usecase.NewAuctionUseCase(auctionRepo, bidRepo)

	out, err := uc.FindWinningBidByAuctionId(context.Background(), id)
	require.Nil(t, err)
	require.Equal(t, id, out.Auction.Id)
	require.NotNil(t, out.Bid)
	require.Equal(t, float64(500), out.Bid.Amount)
}

func TestFindWinningBidByAuctionId_AuctionError(t *testing.T) {
	t.Parallel()

	auctionRepo := &fakeAuctionRepo{byIdErr: internal_error.NewNotFoundError("missing")}
	uc := auction_usecase.NewAuctionUseCase(auctionRepo, &fakeBidRepo{})

	out, err := uc.FindWinningBidByAuctionId(context.Background(), uuid.NewString())
	require.NotNil(t, err)
	require.Nil(t, out)
}

// TestFindWinningBidByAuctionId_BidErrorFailsOpen valida o comportamento fail-open:
// se a busca do lance vencedor falha, retorna o auction com Bid nil e sem erro.
func TestFindWinningBidByAuctionId_BidErrorFailsOpen(t *testing.T) {
	t.Parallel()

	id := uuid.NewString()
	auctionRepo := &fakeAuctionRepo{byId: &auction_entity.Auction{Id: id, Status: auction_entity.Completed}}
	bidRepo := &fakeBidRepo{winningErr: internal_error.NewInternalServerError("no winner")}
	uc := auction_usecase.NewAuctionUseCase(auctionRepo, bidRepo)

	out, err := uc.FindWinningBidByAuctionId(context.Background(), id)
	require.Nil(t, err)
	require.Equal(t, id, out.Auction.Id)
	require.Nil(t, out.Bid)
}
