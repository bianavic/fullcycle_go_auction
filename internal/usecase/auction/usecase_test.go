package auction_test

import (
	"context"
	"testing"
	"time"

	"fullcycle-auction_go/internal/entity/auction"
	"fullcycle-auction_go/internal/entity/bid"
	"fullcycle-auction_go/internal/internal_error"
	auctionuc "fullcycle-auction_go/internal/usecase/auction"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type fakeAuctionRepo struct {
	created   []*auction.Auction
	createErr *internal_error.InternalError

	byID    *auction.Auction
	byIDErr *internal_error.InternalError

	list    []auction.Auction
	listErr *internal_error.InternalError
}

func (f *fakeAuctionRepo) CreateAuction(ctx context.Context, a *auction.Auction) *internal_error.InternalError {
	f.created = append(f.created, a)
	return f.createErr
}

func (f *fakeAuctionRepo) FindAuctions(ctx context.Context, status auction.AuctionStatus, category, productName string) ([]auction.Auction, *internal_error.InternalError) {
	return f.list, f.listErr
}

func (f *fakeAuctionRepo) FindAuctionByID(ctx context.Context, id string) (*auction.Auction, *internal_error.InternalError) {
	return f.byID, f.byIDErr
}

// fakeBidRepo é um stub de bid.BidRepository; só o caminho do lance
// vencedor é exercitado por estes testes.
type fakeBidRepo struct {
	winning    *bid.Bid
	winningErr *internal_error.InternalError
}

func (f *fakeBidRepo) CreateBid(ctx context.Context, bids []bid.Bid) *internal_error.InternalError {
	return nil
}

func (f *fakeBidRepo) FindBidByAuctionID(ctx context.Context, auctionID string) ([]bid.Bid, *internal_error.InternalError) {
	return nil, nil
}

func (f *fakeBidRepo) FindWinningBidByAuctionID(ctx context.Context, auctionID string) (*bid.Bid, *internal_error.InternalError) {
	return f.winning, f.winningErr
}

func validInput() auctionuc.AuctionInputDTO {
	return auctionuc.AuctionInputDTO{
		ProductName: "Clock",
		Category:    "Decor",
		Description: "A long enough description",
		Condition:   1,
	}
}

func TestCreateAuction(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		auctionRepo := &fakeAuctionRepo{}
		uc := auctionuc.New(auctionRepo, &fakeBidRepo{})

		err := uc.CreateAuction(context.Background(), validInput())
		require.Nil(t, err)
		require.Len(t, auctionRepo.created, 1)
		require.Equal(t, "Clock", auctionRepo.created[0].ProductName)
	})

	t.Run("validation error", func(t *testing.T) {
		t.Parallel()
		auctionRepo := &fakeAuctionRepo{}
		uc := auctionuc.New(auctionRepo, &fakeBidRepo{})

		input := validInput()
		input.ProductName = "C"

		err := uc.CreateAuction(context.Background(), input)
		require.NotNil(t, err)
		require.Equal(t, "bad_request", err.Err)
		require.Empty(t, auctionRepo.created)
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()
		auctionRepo := &fakeAuctionRepo{createErr: internal_error.NewInternalServerError("unexpected error")}
		uc := auctionuc.New(auctionRepo, &fakeBidRepo{})

		err := uc.CreateAuction(context.Background(), validInput())
		require.NotNil(t, err)
		require.Equal(t, "internal_server_error", err.Err)
	})
}

func TestFindAuctionByID(t *testing.T) {
	t.Parallel()

	t.Run("maps fields", func(t *testing.T) {
		t.Parallel()
		id := uuid.NewString()
		ts := time.Now()
		auctionRepo := &fakeAuctionRepo{byID: &auction.Auction{
			ID: id, ProductName: "Clock", Category: "Decor",
			Description: "desc", Condition: auction.New,
			Status: auction.Active, Timestamp: ts,
		}}
		uc := auctionuc.New(auctionRepo, &fakeBidRepo{})

		out, err := uc.FindAuctionByID(context.Background(), id)
		require.Nil(t, err)
		require.Equal(t, id, out.ID)
		require.Equal(t, "Clock", out.ProductName)
		require.Equal(t, auctionuc.AuctionStatus(auction.Active), out.Status)
		require.Equal(t, ts, out.Timestamp)
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()
		auctionRepo := &fakeAuctionRepo{byIDErr: internal_error.NewNotFoundError("missing")}
		uc := auctionuc.New(auctionRepo, &fakeBidRepo{})

		out, err := uc.FindAuctionByID(context.Background(), uuid.NewString())
		require.NotNil(t, err)
		require.Nil(t, out)
	})
}

func TestFindAuctions(t *testing.T) {
	t.Parallel()

	t.Run("maps fields", func(t *testing.T) {
		t.Parallel()
		auctionRepo := &fakeAuctionRepo{list: []auction.Auction{
			{ID: uuid.NewString(), ProductName: "A", Status: auction.Active},
			{ID: uuid.NewString(), ProductName: "B", Status: auction.Completed},
		}}
		uc := auctionuc.New(auctionRepo, &fakeBidRepo{})

		out, err := uc.FindAuctions(context.Background(), 0, "", "")
		require.Nil(t, err)
		require.Len(t, out, 2)
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()
		auctionRepo := &fakeAuctionRepo{listErr: internal_error.NewInternalServerError("unexpected error")}
		uc := auctionuc.New(auctionRepo, &fakeBidRepo{})

		out, err := uc.FindAuctions(context.Background(), 0, "", "")
		require.NotNil(t, err)
		require.Nil(t, out)
	})
}

func TestFindWinningBidByAuctionID(t *testing.T) {
	t.Parallel()

	t.Run("returns bid", func(t *testing.T) {
		t.Parallel()
		id := uuid.NewString()
		auctionRepo := &fakeAuctionRepo{byID: &auction.Auction{ID: id, Status: auction.Completed}}
		bidRepo := &fakeBidRepo{winning: &bid.Bid{ID: uuid.NewString(), AuctionID: id, Amount: 500}}
		uc := auctionuc.New(auctionRepo, bidRepo)

		out, err := uc.FindWinningBidByAuctionID(context.Background(), id)
		require.Nil(t, err)
		require.Equal(t, id, out.Auction.ID)
		require.NotNil(t, out.Bid)
		require.Equal(t, float64(500), out.Bid.Amount)
	})

	t.Run("auction error", func(t *testing.T) {
		t.Parallel()
		auctionRepo := &fakeAuctionRepo{byIDErr: internal_error.NewNotFoundError("missing")}
		uc := auctionuc.New(auctionRepo, &fakeBidRepo{})

		out, err := uc.FindWinningBidByAuctionID(context.Background(), uuid.NewString())
		require.NotNil(t, err)
		require.Nil(t, out)
	})

	// fail-open: se a busca do lance vencedor falha, retorna o auction com Bid nil e sem erro.
	t.Run("bid error fails open", func(t *testing.T) {
		t.Parallel()
		id := uuid.NewString()
		auctionRepo := &fakeAuctionRepo{byID: &auction.Auction{ID: id, Status: auction.Completed}}
		bidRepo := &fakeBidRepo{winningErr: internal_error.NewInternalServerError("no winner")}
		uc := auctionuc.New(auctionRepo, bidRepo)

		out, err := uc.FindWinningBidByAuctionID(context.Background(), id)
		require.Nil(t, err)
		require.Equal(t, id, out.Auction.ID)
		require.Nil(t, out.Bid)
	})
}
