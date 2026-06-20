package bid_usecase_test

import (
	"context"
	"testing"
	"time"

	"fullcycle-auction_go/internal/entity/bid_entity"
	"fullcycle-auction_go/internal/internal_error"
	"fullcycle-auction_go/internal/usecase/bid_usecase"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// TestFindBidByAuctionId_MapsFields valida o mapeamento Bid -> BidOutputDTO.
func TestFindBidByAuctionId_MapsFields(t *testing.T) {
	auctionId := uuid.NewString()
	bidId := uuid.NewString()
	userId := uuid.NewString()
	ts := time.Now()

	repo := &fakeBidRepo{findBids: []bid_entity.Bid{
		{Id: bidId, UserId: userId, AuctionId: auctionId, Amount: 150, Timestamp: ts},
	}}
	uc := bid_usecase.NewBidUseCase(repo)

	out, err := uc.FindBidByAuctionId(context.Background(), auctionId)
	require.Nil(t, err)
	require.Len(t, out, 1)
	require.Equal(t, bidId, out[0].Id)
	require.Equal(t, userId, out[0].UserId)
	require.Equal(t, auctionId, out[0].AuctionId)
	require.Equal(t, float64(150), out[0].Amount)
	require.Equal(t, ts, out[0].Timestamp)
}

// TestFindBidByAuctionId_RepositoryError valida a propagação de erro do repositório.
func TestFindBidByAuctionId_RepositoryError(t *testing.T) {
	repo := &fakeBidRepo{findErr: internal_error.NewInternalServerError("boom")}
	uc := bid_usecase.NewBidUseCase(repo)

	out, err := uc.FindBidByAuctionId(context.Background(), uuid.NewString())
	require.NotNil(t, err)
	require.Nil(t, out)
}

// TestFindWinningBidByAuctionId_MapsFields valida o mapeamento do lance vencedor.
func TestFindWinningBidByAuctionId_MapsFields(t *testing.T) {
	auctionId := uuid.NewString()
	bidId := uuid.NewString()
	userId := uuid.NewString()
	ts := time.Now()

	repo := &fakeBidRepo{winning: &bid_entity.Bid{
		Id: bidId, UserId: userId, AuctionId: auctionId, Amount: 999, Timestamp: ts,
	}}
	uc := bid_usecase.NewBidUseCase(repo)

	out, err := uc.FindWinningBidByAuctionId(context.Background(), auctionId)
	require.Nil(t, err)
	require.Equal(t, bidId, out.Id)
	require.Equal(t, userId, out.UserId)
	require.Equal(t, auctionId, out.AuctionId)
	require.Equal(t, float64(999), out.Amount)
	require.Equal(t, ts, out.Timestamp)
}

// TestFindWinningBidByAuctionId_RepositoryError valida a propagação de erro.
func TestFindWinningBidByAuctionId_RepositoryError(t *testing.T) {
	repo := &fakeBidRepo{winningErr: internal_error.NewInternalServerError("boom")}
	uc := bid_usecase.NewBidUseCase(repo)

	out, err := uc.FindWinningBidByAuctionId(context.Background(), uuid.NewString())
	require.NotNil(t, err)
	require.Nil(t, out)
}
