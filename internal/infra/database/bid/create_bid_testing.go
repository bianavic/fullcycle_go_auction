//go:build integration

package bid

import (
	"context"
)

// InsertBidForTest insere um bid diretamente na coleção, sem passar pela
// validação de status do leilão feita por CreateBid. Usado por testes de
// integração que precisam de bids pré-existentes para exercitar as buscas.
func (bd *BidRepository) InsertBidForTest(
	ctx context.Context, id, userID, auctionID string, amount float64, timestamp int64) error {
	bidMongo := &BidMongo{
		ID:        id,
		UserID:    userID,
		AuctionID: auctionID,
		Amount:    amount,
		Timestamp: timestamp,
	}

	_, err := bd.Collection.InsertOne(ctx, bidMongo)
	return err
}
