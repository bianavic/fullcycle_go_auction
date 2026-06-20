//go:build integration

package bid

import (
	"context"
)

// InsertBidForTest insere um bid diretamente na coleção, sem passar pela
// validação de status do leilão feita por CreateBid. Usado por testes de
// integração que precisam de bids pré-existentes para exercitar as buscas.
func (bd *BidRepository) InsertBidForTest(
	ctx context.Context, id, userId, auctionId string, amount float64, timestamp int64) error {
	bidEntityMongo := &BidEntityMongo{
		Id:        id,
		UserId:    userId,
		AuctionId: auctionId,
		Amount:    amount,
		Timestamp: timestamp,
	}

	_, err := bd.Collection.InsertOne(ctx, bidEntityMongo)
	return err
}
