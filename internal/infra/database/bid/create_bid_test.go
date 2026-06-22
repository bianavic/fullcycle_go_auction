//go:build integration

package bid_test

import (
	"context"
	"testing"
	"time"

	"fullcycle-auction_go/internal/entity/auction"
	"fullcycle-auction_go/internal/entity/bid"
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

	auctionID := uuid.NewString()
	require.NoError(t, auctionRepo.InsertAuctionForTest(ctx, auctionID,
		"Live Auction", "Cat", "an active auction for integration",
		auction.New, auction.Active, time.Now().Unix()))

	bid1, errBid := bid.CreateBid(uuid.NewString(), auctionID, 100)
	require.Nil(t, errBid)
	bid2, errBid := bid.CreateBid(uuid.NewString(), auctionID, 200)
	require.Nil(t, errBid)

	require.Nil(t, bidRepo.CreateBid(ctx, []bid.Bid{*bid1, *bid2}))

	bids, err := bidRepo.FindBidByAuctionID(ctx, auctionID)
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

	auctionID := uuid.NewString()
	require.NoError(t, auctionRepo.InsertAuctionForTest(ctx, auctionID,
		"Closed Auction", "Cat", "a completed auction for integration",
		auction.New, auction.Completed, time.Now().Unix()))

	bid, errBid := bid.CreateBid(uuid.NewString(), auctionID, 100)
	require.Nil(t, errBid)

	require.Nil(t, bidRepo.CreateBid(ctx, []bid.Bid{*bid}))

	bids, err := bidRepo.FindBidByAuctionID(ctx, auctionID)
	require.Nil(t, err)
	require.Empty(t, bids)
}
