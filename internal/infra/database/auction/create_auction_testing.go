//go:build integration

package auction

import (
	"context"

	"fullcycle-auction_go/internal/entity/auction_entity"
)

// InsertAuctionForTest insere um leilão com campos arbitrários diretamente na
// coleção, sem disparar scheduleAuctionClose nem a validação da entidade. Permite
// que testes de integração no pacote externo montem cenários (status, categoria,
// productName) sem acesso à struct interna AuctionEntityMongo.
func (ar *AuctionRepository) InsertAuctionForTest(
	ctx context.Context,
	id, productName, category, description string,
	condition auction_entity.ProductCondition,
	status auction_entity.AuctionStatus,
	timestamp int64,
) error {
	auctionEntityMongo := &AuctionEntityMongo{
		Id:          id,
		ProductName: productName,
		Category:    category,
		Description: description,
		Condition:   condition,
		Status:      status,
		Timestamp:   timestamp,
	}

	_, err := ar.Collection.InsertOne(ctx, auctionEntityMongo)
	return err
}

// InsertExpiredAuctionForTest insere um leilão Active com timestamp no passado
// diretamente na coleção, sem disparar scheduleAuctionClose. Usado por testes de
// integração que precisam de um leilão já vencido para exercitar o monitor de
// fechamento sem depender da goroutine agendada.
func (ar *AuctionRepository) InsertExpiredAuctionForTest(
	ctx context.Context, id string, pastTimestamp int64) error {
	return ar.InsertAuctionForTest(ctx, id,
		"test product", "test category", "test description for integration",
		auction_entity.New, auction_entity.Active, pastTimestamp)
}
