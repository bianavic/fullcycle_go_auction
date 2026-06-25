//go:build integration

package auction

import (
	"context"

	"fullcycle-auction_go/internal/entity/auction"
)

// InsertAuctionForTest insere um leilão diretamente na coleção, ignorando validações e
// rotinas de fechamento.
func (r *Repository) InsertAuctionForTest(
	ctx context.Context,
	id, productName, category, description string,
	condition auction.ProductCondition,
	status auction.Status,
	timestamp int64,
) error {
	doc := &document{
		ID:          id,
		ProductName: productName,
		Category:    category,
		Description: description,
		Condition:   condition,
		Status:      status,
		Timestamp:   timestamp,
	}

	_, err := r.Collection.InsertOne(ctx, doc)
	return err
}

// InsertExpiredAuctionForTest insere um leilão expirado para testes do monitor
// de fechamento.
func (r *Repository) InsertExpiredAuctionForTest(
	ctx context.Context, id string, pastTimestamp int64) error {
	return r.InsertAuctionForTest(ctx, id,
		"test product", "test category", "test description for integration",
		auction.New, auction.Active, pastTimestamp)
}
