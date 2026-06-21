//go:build integration

package auction_test

import (
	"context"
	"testing"
	"time"

	"fullcycle-auction_go/internal/entity/auction_entity"
	"fullcycle-auction_go/internal/infra/database/auction"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// TestFindAuctionById_Found insere um leilão e o recupera por ID exato.
func TestFindAuctionById_Found(t *testing.T) {
	t.Parallel()

	db := setupMongo(t)
	ctx := context.Background()
	repo := auction.NewAuctionRepository(ctx, db)

	id := uuid.NewString()
	require.NoError(t, repo.InsertAuctionForTest(ctx, id,
		"Vintage Clock", "Decor", "A beautiful vintage wall clock",
		auction_entity.New, auction_entity.Active, time.Now().Unix()))

	found, err := repo.FindAuctionById(ctx, id)
	require.Nil(t, err)
	require.Equal(t, id, found.Id)
	require.Equal(t, "Vintage Clock", found.ProductName)
	require.Equal(t, "Decor", found.Category)
	require.Equal(t, auction_entity.Active, found.Status)
}

// TestFindAuctionById_NotFound confirma que um ID inexistente retorna erro.
func TestFindAuctionById_NotFound(t *testing.T) {
	t.Parallel()

	db := setupMongo(t)
	ctx := context.Background()
	repo := auction.NewAuctionRepository(ctx, db)

	found, err := repo.FindAuctionById(ctx, uuid.NewString())
	require.NotNil(t, err)
	require.Nil(t, found)
}

// TestFindAuctions_ByStatus valida o filtro de status. Como Active == 0 e o filtro
// só é aplicado quando status != 0, apenas o filtro por Completed é seletivo.
func TestFindAuctions_ByStatus(t *testing.T) {
	t.Parallel()

	db := setupMongo(t)
	ctx := context.Background()
	repo := auction.NewAuctionRepository(ctx, db)
	ts := time.Now().Unix()

	activeID := uuid.NewString()
	completedID := uuid.NewString()
	require.NoError(t, repo.InsertAuctionForTest(ctx, activeID,
		"Active Item", "Cat", "an active auction for integration",
		auction_entity.New, auction_entity.Active, ts))
	require.NoError(t, repo.InsertAuctionForTest(ctx, completedID,
		"Completed Item", "Cat", "a completed auction for integration",
		auction_entity.New, auction_entity.Completed, ts))

	completed, err := repo.FindAuctions(ctx, auction_entity.Completed, "", "")
	require.Nil(t, err)
	require.Len(t, completed, 1)
	require.Equal(t, completedID, completed[0].Id)
}

// TestFindAuctions_ByCategory valida o filtro exato por categoria.
func TestFindAuctions_ByCategory(t *testing.T) {
	t.Parallel()

	db := setupMongo(t)
	ctx := context.Background()
	repo := auction.NewAuctionRepository(ctx, db)
	ts := time.Now().Unix()

	artID := uuid.NewString()
	require.NoError(t, repo.InsertAuctionForTest(ctx, artID,
		"Oil Painting", "Art", "an art auction for integration",
		auction_entity.New, auction_entity.Active, ts))
	require.NoError(t, repo.InsertAuctionForTest(ctx, uuid.NewString(),
		"Vintage Clock", "Decor", "a decor auction for integration",
		auction_entity.New, auction_entity.Active, ts))

	result, err := repo.FindAuctions(ctx, 0, "Art", "")
	require.Nil(t, err)
	require.Len(t, result, 1)
	require.Equal(t, artID, result[0].Id)
}

// TestFindAuctions_ByProductName valida o filtro regex case-insensitive sobre
// product_name. Antes do fix do bug BSON (chave "productName"), o filtro não
// encontrava o campo e retornava todos os documentos — este teste falharia.
func TestFindAuctions_ByProductName(t *testing.T) {
	t.Parallel()

	db := setupMongo(t)
	ctx := context.Background()
	repo := auction.NewAuctionRepository(ctx, db)
	ts := time.Now().Unix()

	clockID := uuid.NewString()
	require.NoError(t, repo.InsertAuctionForTest(ctx, clockID,
		"Vintage Clock", "Decor", "a clock auction for integration",
		auction_entity.New, auction_entity.Active, ts))
	require.NoError(t, repo.InsertAuctionForTest(ctx, uuid.NewString(),
		"Oil Painting", "Art", "a painting auction for integration",
		auction_entity.New, auction_entity.Active, ts))

	result, err := repo.FindAuctions(ctx, 0, "", "clock")
	require.Nil(t, err)
	require.Len(t, result, 1)
	require.Equal(t, clockID, result[0].Id)
}

// TestFindAuctions_EmptyFilter confirma que status=0 sem categoria nem productName
// retorna todos os leilões.
func TestFindAuctions_EmptyFilter(t *testing.T) {
	t.Parallel()

	db := setupMongo(t)
	ctx := context.Background()
	repo := auction.NewAuctionRepository(ctx, db)
	ts := time.Now().Unix()

	for i := 0; i < 3; i++ {
		require.NoError(t, repo.InsertAuctionForTest(ctx, uuid.NewString(),
			"Item", "Cat", "an auction for integration tests",
			auction_entity.New, auction_entity.Active, ts))
	}

	result, err := repo.FindAuctions(ctx, 0, "", "")
	require.Nil(t, err)
	require.Len(t, result, 3)
}
