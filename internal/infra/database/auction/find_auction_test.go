//go:build integration

package auction_test

import (
	"context"
	"testing"
	"time"

	"fullcycle-auction_go/internal/entity/auction"
	auctiondb "fullcycle-auction_go/internal/infra/database/auction"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestFindAuctionByID(t *testing.T) {
	t.Parallel()

	t.Run("found", func(t *testing.T) {
		t.Parallel()
		db := setupMongo(t)
		ctx := context.Background()
		repo := auctiondb.New(ctx, db)

		id := uuid.NewString()
		require.NoError(t, repo.InsertAuctionForTest(ctx, id,
			"Vintage Clock", "Decor", "A beautiful vintage wall clock",
			auction.New, auction.Active, time.Now().Unix()))

		found, err := repo.FindAuctionByID(ctx, id)
		require.Nil(t, err)
		require.Equal(t, id, found.ID)
		require.Equal(t, "Vintage Clock", found.ProductName)
		require.Equal(t, "Decor", found.Category)
		require.Equal(t, auction.Active, found.Status)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		db := setupMongo(t)
		ctx := context.Background()
		repo := auctiondb.New(ctx, db)

		found, err := repo.FindAuctionByID(ctx, uuid.NewString())
		require.NotNil(t, err)
		require.Nil(t, found)
	})
}

func TestFindAuctions(t *testing.T) {
	t.Parallel()

	// Como Active == 0 e o filtro só é aplicado quando status != 0, apenas o filtro
	// por Completed é seletivo.
	t.Run("by status", func(t *testing.T) {
		t.Parallel()
		db := setupMongo(t)
		ctx := context.Background()
		repo := auctiondb.New(ctx, db)
		ts := time.Now().Unix()

		activeID := uuid.NewString()
		completedID := uuid.NewString()
		require.NoError(t, repo.InsertAuctionForTest(ctx, activeID,
			"Active Item", "Cat", "an active auction for integration",
			auction.New, auction.Active, ts))
		require.NoError(t, repo.InsertAuctionForTest(ctx, completedID,
			"Completed Item", "Cat", "a completed auction for integration",
			auction.New, auction.Completed, ts))

		completed, err := repo.FindAuctions(ctx, auction.Completed, "", "")
		require.Nil(t, err)
		require.Len(t, completed, 1)
		require.Equal(t, completedID, completed[0].ID)
	})

	t.Run("by category", func(t *testing.T) {
		t.Parallel()
		db := setupMongo(t)
		ctx := context.Background()
		repo := auctiondb.New(ctx, db)
		ts := time.Now().Unix()

		artID := uuid.NewString()
		require.NoError(t, repo.InsertAuctionForTest(ctx, artID,
			"Oil Painting", "Art", "an art auction for integration",
			auction.New, auction.Active, ts))
		require.NoError(t, repo.InsertAuctionForTest(ctx, uuid.NewString(),
			"Vintage Clock", "Decor", "a decor auction for integration",
			auction.New, auction.Active, ts))

		result, err := repo.FindAuctions(ctx, 0, "Art", "")
		require.Nil(t, err)
		require.Len(t, result, 1)
		require.Equal(t, artID, result[0].ID)
	})

	// Valida o filtro regex case-insensitive sobre product_name. Antes do fix do bug
	// BSON (chave "productName"), o filtro não encontrava o campo e retornava todos os
	// documentos — este teste falharia.
	t.Run("by product name", func(t *testing.T) {
		t.Parallel()
		db := setupMongo(t)
		ctx := context.Background()
		repo := auctiondb.New(ctx, db)
		ts := time.Now().Unix()

		clockID := uuid.NewString()
		require.NoError(t, repo.InsertAuctionForTest(ctx, clockID,
			"Vintage Clock", "Decor", "a clock auction for integration",
			auction.New, auction.Active, ts))
		require.NoError(t, repo.InsertAuctionForTest(ctx, uuid.NewString(),
			"Oil Painting", "Art", "a painting auction for integration",
			auction.New, auction.Active, ts))

		result, err := repo.FindAuctions(ctx, 0, "", "clock")
		require.Nil(t, err)
		require.Len(t, result, 1)
		require.Equal(t, clockID, result[0].ID)
	})

	t.Run("empty filter returns all", func(t *testing.T) {
		t.Parallel()
		db := setupMongo(t)
		ctx := context.Background()
		repo := auctiondb.New(ctx, db)
		ts := time.Now().Unix()

		for i := 0; i < 3; i++ {
			require.NoError(t, repo.InsertAuctionForTest(ctx, uuid.NewString(),
				"Item", "Cat", "an auction for integration tests",
				auction.New, auction.Active, ts))
		}

		result, err := repo.FindAuctions(ctx, 0, "", "")
		require.Nil(t, err)
		require.Len(t, result, 3)
	})
}