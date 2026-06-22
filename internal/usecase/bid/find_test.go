package bid_test

import (
	"context"
	"testing"
	"time"

	"fullcycle-auction_go/internal/entity/bid"
	"fullcycle-auction_go/internal/internal_error"
	biduc "fullcycle-auction_go/internal/usecase/bid"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// TestFindBidByAuctionID_MapsFields valida o mapeamento Bid -> BidOutputDTO.
func TestFindBidByAuctionID_MapsFields(t *testing.T) {
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
}

// TestFindBidByAuctionID_RepositoryError valida a propagação de erro do repositório.
func TestFindBidByAuctionID_RepositoryError(t *testing.T) {
	repo := &fakeBidRepo{findErr: internal_error.NewInternalServerError("unexpected error")}
	uc := biduc.New(repo)

	out, err := uc.FindBidByAuctionID(context.Background(), uuid.NewString())
	require.NotNil(t, err)
	require.Nil(t, out)
}

// TestFindWinningBidByAuctionID_MapsFields valida o mapeamento do lance vencedor.
func TestFindWinningBidByAuctionID_MapsFields(t *testing.T) {
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
}

// TestFindWinningBidByAuctionID_RepositoryError valida a propagação de erro.
func TestFindWinningBidByAuctionID_RepositoryError(t *testing.T) {
	repo := &fakeBidRepo{winningErr: internal_error.NewInternalServerError("unexpected error")}
	uc := biduc.New(repo)

	out, err := uc.FindWinningBidByAuctionID(context.Background(), uuid.NewString())
	require.NotNil(t, err)
	require.Nil(t, out)
}
