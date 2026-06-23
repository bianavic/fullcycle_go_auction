package bid_test

import (
	"context"
	"testing"
	"time"

	"fullcycle-auction_go/internal/entity/bid"
	"fullcycle-auction_go/internal/apperr"
	biduc "fullcycle-auction_go/internal/usecase/bid"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestFindBidByAuctionID(t *testing.T) {
	t.Parallel()

	t.Run("maps fields", func(t *testing.T) {
		auctionID := uuid.NewString()
		bidID := uuid.NewString()
		userID := uuid.NewString()
		ts := time.Now()

		repo := &fakeBidRepo{findBids: []bid.Bid{
			{ID: bidID, UserID: userID, AuctionID: auctionID, Amount: 150, Timestamp: ts},
		}}
		uc := biduc.New(repo)

		out, err := uc.FindBidByAuctionID(context.Background(), auctionID)
		require.Nil(t, err)
		require.Len(t, out, 1)
		require.Equal(t, bidID, out[0].ID)
		require.Equal(t, userID, out[0].UserID)
		require.Equal(t, auctionID, out[0].AuctionID)
		require.Equal(t, float64(150), out[0].Amount)
		require.Equal(t, ts, out[0].Timestamp)
	})

	t.Run("repository error", func(t *testing.T) {
		repo := &fakeBidRepo{findErr: apperr.NewInternalServerError("unexpected error")}
		uc := biduc.New(repo)

		out, err := uc.FindBidByAuctionID(context.Background(), uuid.NewString())
		require.NotNil(t, err)
		require.Nil(t, out)
	})
}

func TestFindWinningBidByAuctionID(t *testing.T) {
	t.Parallel()

	t.Run("maps fields", func(t *testing.T) {
		auctionID := uuid.NewString()
		bidID := uuid.NewString()
		userID := uuid.NewString()
		ts := time.Now()

		repo := &fakeBidRepo{winning: &bid.Bid{
			ID: bidID, UserID: userID, AuctionID: auctionID, Amount: 999, Timestamp: ts,
		}}
		uc := biduc.New(repo)

		out, err := uc.FindWinningBidByAuctionID(context.Background(), auctionID)
		require.Nil(t, err)
		require.Equal(t, bidID, out.ID)
		require.Equal(t, userID, out.UserID)
		require.Equal(t, auctionID, out.AuctionID)
		require.Equal(t, float64(999), out.Amount)
		require.Equal(t, ts, out.Timestamp)
	})

	t.Run("repository error", func(t *testing.T) {
		repo := &fakeBidRepo{winningErr: apperr.NewInternalServerError("unexpected error")}
		uc := biduc.New(repo)

		out, err := uc.FindWinningBidByAuctionID(context.Background(), uuid.NewString())
		require.NotNil(t, err)
		require.Nil(t, out)
	})
}
