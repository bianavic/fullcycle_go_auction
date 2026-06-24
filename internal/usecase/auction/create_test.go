package auction_test

import (
	"context"
	"fullcycle-auction_go/internal/apperr"
	"fullcycle-auction_go/internal/entity/auction"
	"fullcycle-auction_go/internal/entity/bid"
	auctionuc "fullcycle-auction_go/internal/usecase/auction"
	"testing"

	"github.com/stretchr/testify/require"
)

func (f *fakeAuctionRepo) Create(ctx context.Context, a *auction.Auction) *apperr.InternalError {
	f.created = append(f.created, a)
	return f.createErr
}

func (f *fakeBidRepo) Create(ctx context.Context, bids []bid.Bid) *apperr.InternalError {
	return nil
}

func validInput() auctionuc.InputDTO {
	return auctionuc.InputDTO{
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
		auctionRepo := &fakeAuctionRepo{createErr: apperr.NewInternalServerError("unexpected error")}
		uc := auctionuc.New(auctionRepo, &fakeBidRepo{})

		err := uc.CreateAuction(context.Background(), validInput())
		require.NotNil(t, err)
		require.Equal(t, "internal_server_error", err.Err)
	})
}
