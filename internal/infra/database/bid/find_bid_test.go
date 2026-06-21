//go:build integration

package bid_test

import (
	"context"
	"testing"
	"time"

	"fullcycle-auction_go/internal/infra/database/auction"
	"fullcycle-auction_go/internal/infra/database/bid"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// setupMongo sobe um MongoDB efêmero via Testcontainers e devolve um *mongo.Database
// pronto para uso. A limpeza é registrada com t.Cleanup.
func setupMongo(t *testing.T) *mongo.Database {
	t.Helper()
	ctx := context.Background()

	container, err := mongodb.Run(ctx, "mongo:7")
	if err != nil {
		t.Fatalf("failed to start mongodb container: %v", err)
	}
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			t.Logf("failed to terminate mongodb container: %v", err)
		}
	})

	uri, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatalf("failed to connect to mongodb: %v", err)
	}
	t.Cleanup(func() {
		_ = client.Disconnect(ctx)
	})

	if err := client.Ping(ctx, nil); err != nil {
		t.Fatalf("failed to ping mongodb: %v", err)
	}

	return client.Database("auctions_test")
}

// newBidRepository monta um BidRepository ligado ao db informado, com o
// AuctionRepository exigido pelo construtor.
func newBidRepository(t *testing.T, db *mongo.Database) *bid.BidRepository {
	t.Helper()
	return bid.NewBidRepository(db, auction.NewAuctionRepository(context.Background(), db))
}

// TestFindBidByAuctionId_ReturnsBids valida que apenas os bids do auction filtrado
// são retornados.
func TestFindBidByAuctionId_ReturnsBids(t *testing.T) {
	t.Parallel()

	db := setupMongo(t)
	repo := newBidRepository(t, db)
	ctx := context.Background()
	ts := time.Now().Unix()

	auctionId := uuid.NewString()
	otherAuctionId := uuid.NewString()
	require.NoError(t, repo.InsertBidForTest(ctx, uuid.NewString(), uuid.NewString(), auctionId, 100, ts))
	require.NoError(t, repo.InsertBidForTest(ctx, uuid.NewString(), uuid.NewString(), auctionId, 200, ts))
	require.NoError(t, repo.InsertBidForTest(ctx, uuid.NewString(), uuid.NewString(), otherAuctionId, 300, ts))

	bids, err := repo.FindBidByAuctionId(ctx, auctionId)
	require.Nil(t, err)
	require.Len(t, bids, 2)
}

// TestFindBidByAuctionId_EmptyWhenNone confirma que um auction sem bids retorna
// slice vazio sem erro.
func TestFindBidByAuctionId_EmptyWhenNone(t *testing.T) {
	t.Parallel()

	db := setupMongo(t)
	repo := newBidRepository(t, db)
	ctx := context.Background()

	bids, err := repo.FindBidByAuctionId(ctx, uuid.NewString())
	require.Nil(t, err)
	require.Empty(t, bids)
}

// TestFindWinningBidByAuctionId_ReturnsHighestAmount valida a ordenação descendente
// por amount: o vencedor é o maior lance.
func TestFindWinningBidByAuctionId_ReturnsHighestAmount(t *testing.T) {
	t.Parallel()

	db := setupMongo(t)
	repo := newBidRepository(t, db)
	ctx := context.Background()
	ts := time.Now().Unix()

	auctionId := uuid.NewString()
	winnerId := uuid.NewString()
	require.NoError(t, repo.InsertBidForTest(ctx, uuid.NewString(), uuid.NewString(), auctionId, 100, ts))
	require.NoError(t, repo.InsertBidForTest(ctx, winnerId, uuid.NewString(), auctionId, 300, ts))
	require.NoError(t, repo.InsertBidForTest(ctx, uuid.NewString(), uuid.NewString(), auctionId, 200, ts))

	winner, err := repo.FindWinningBidByAuctionId(ctx, auctionId)
	require.Nil(t, err)
	require.Equal(t, winnerId, winner.Id)
	require.Equal(t, float64(300), winner.Amount)
}

// TestFindWinningBidByAuctionId_NotFound confirma que um auction sem bids retorna
// erro.
func TestFindWinningBidByAuctionId_NotFound(t *testing.T) {
	t.Parallel()

	db := setupMongo(t)
	repo := newBidRepository(t, db)
	ctx := context.Background()

	winner, err := repo.FindWinningBidByAuctionId(ctx, uuid.NewString())
	require.NotNil(t, err)
	require.Nil(t, winner)
}
