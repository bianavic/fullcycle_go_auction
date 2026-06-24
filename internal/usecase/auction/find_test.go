package auction_test

import (
	"context"
	"fullcycle-auction_go/internal/apperr"
	"fullcycle-auction_go/internal/entity/auction"
	"fullcycle-auction_go/internal/entity/bid"
	auctionuc "fullcycle-auction_go/internal/usecase/auction"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func (f *fakeAuctionRepo) FindAll(ctx context.Context, status auction.Status, category, productName string) ([]auction.Auction, *apperr.InternalError) {
	return f.list, f.listErr
}

func (f *fakeAuctionRepo) FindByID(ctx context.Context, id string) (*auction.Auction, *apperr.InternalError) {
	return f.byID, f.byIDErr
}

func (f *fakeBidRepo) FindByAuctionID(ctx context.Context, auctionID string) ([]bid.Bid, *apperr.InternalError) {
	return nil, nil
}

func (f *fakeBidRepo) FindWinningByAuctionID(ctx context.Context, auctionID string) (*bid.Bid, *apperr.InternalError) {
	return f.winning, f.winningErr
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
		auctionRepo := &fakeAuctionRepo{byIDErr: apperr.NewNotFoundError("missing")}
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
		auctionRepo := &fakeAuctionRepo{listErr: apperr.NewInternalServerError("unexpected error")}
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
		auctionRepo := &fakeAuctionRepo{byIDErr: apperr.NewNotFoundError("missing")}
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
		bidRepo := &fakeBidRepo{winningErr: apperr.NewInternalServerError("no winner")}
		uc := auctionuc.New(auctionRepo, bidRepo)

		out, err := uc.FindWinningBidByAuctionID(context.Background(), id)
		require.Nil(t, err)
		require.Equal(t, id, out.Auction.ID)
		require.Nil(t, out.Bid)
	})
}
