//go:build integration

package bid

import (
	"context"
)

// InsertBidForTest insere um bid diretamente na coleção para montagem de cenários de teste.
func (bd *Repository) InsertBidForTest(
	ctx context.Context, id, userID, auctionID string, amount float64, timestamp int64) error {
	doc := &document{
		ID:        id,
		UserID:    userID,
		AuctionID: auctionID,
		Amount:    amount,
		Timestamp: timestamp,
	}

	_, err := bd.Collection.InsertOne(ctx, doc)
	return err
}
