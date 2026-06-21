//go:build integration

package bid_test

import (
	"context"
	"testing"
	"time"

	"fullcycle-auction_go/internal/entity/auction_entity"
	"fullcycle-auction_go/internal/entity/bid_entity"
	"fullcycle-auction_go/internal/infra/database/auction"
	"fullcycle-auction_go/internal/infra/database/bid"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// TestCreateBid_InsertsValidBids valida que um lote de bids para um leilão Active é
// persistido.
func TestCreateBid_InsertsValidBids(t *testing.T) {
	t.Parallel()

	db := setupMongo(t)
	ctx := context.Background()
	auctionRepo := auction.NewAuctionRepository(ctx, db)
	bidRepo := bid.NewBidRepository(db, auctionRepo)

	auctionId := uuid.NewString()
	require.NoError(t, auctionRepo.InsertAuctionForTest(ctx, auctionId,
		"Live Auction", "Cat", "an active auction for integration",
		auction_entity.New, auction_entity.Active, time.Now().Unix()))

	bid1, errBid := bid_entity.CreateBid(uuid.NewString(), auctionId, 100)
	require.Nil(t, errBid)
	bid2, errBid := bid_entity.CreateBid(uuid.NewString(), auctionId, 200)
	require.Nil(t, errBid)

	require.Nil(t, bidRepo.CreateBid(ctx, []bid_entity.Bid{*bid1, *bid2}))

	bids, err := bidRepo.FindBidByAuctionId(ctx, auctionId)
	require.Nil(t, err)
	require.Len(t, bids, 2)
}

// TestCreateBid_RejectsCompletedAuction valida que bids para um leilão Completed são
// descartados silenciosamente (não persistidos).
func TestCreateBid_RejectsCompletedAuction(t *testing.T) {
	t.Parallel()

	db := setupMongo(t)
	ctx := context.Background()
	auctionRepo := auction.NewAuctionRepository(ctx, db)
	bidRepo := bid.NewBidRepository(db, auctionRepo)

	auctionId := uuid.NewString()
	require.NoError(t, auctionRepo.InsertAuctionForTest(ctx, auctionId,
		"Closed Auction", "Cat", "a completed auction for integration",
		auction_entity.New, auction_entity.Completed, time.Now().Unix()))

	bidEntity, errBid := bid_entity.CreateBid(uuid.NewString(), auctionId, 100)
	require.Nil(t, errBid)

	require.Nil(t, bidRepo.CreateBid(ctx, []bid_entity.Bid{*bidEntity}))

	bids, err := bidRepo.FindBidByAuctionId(ctx, auctionId)
	require.Nil(t, err)
	require.Empty(t, bids)
}
