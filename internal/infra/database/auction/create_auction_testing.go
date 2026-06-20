//go:build integration

package auction

import (
	"context"

	"fullcycle-auction_go/internal/entity/auction_entity"
)

// InsertExpiredAuctionForTest insere um leilão Active com timestamp no passado
// diretamente na coleção, sem disparar scheduleAuctionClose. Usado por testes de
// integração que precisam de um leilão já vencido para exercitar o monitor de
// fechamento sem depender da goroutine agendada.
func (ar *AuctionRepository) InsertExpiredAuctionForTest(
	ctx context.Context, id string, pastTimestamp int64) error {
	auctionEntityMongo := &AuctionEntityMongo{
		Id:          id,
		ProductName: "test product",
		Category:    "test category",
		Description: "test description for integration",
		Condition:   auction_entity.New,
		Status:      auction_entity.Active,
		Timestamp:   pastTimestamp,
	}

	_, err := ar.Collection.InsertOne(ctx, auctionEntityMongo)
	return err
}