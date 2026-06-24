//go:build integration

package auction

import (
	"context"

	"fullcycle-auction_go/internal/entity/auction"
)

// InsertAuctionForTest insere um leilão com campos arbitrários diretamente na
// coleção, sem disparar scheduleAuctionClose nem a validação da entidade. Permite
// que testes de integração no pacote externo montem cenários (status, categoria,
// productName) sem acesso à struct interna document.
func (ar *Repository) InsertAuctionForTest(
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

	_, err := ar.Collection.InsertOne(ctx, doc)
	return err
}

// InsertExpiredAuctionForTest insere um leilão Active com timestamp no passado
// diretamente na coleção, sem disparar scheduleAuctionClose. Usado por testes de
// integração que precisam de um leilão já vencido para exercitar o monitor de
// fechamento sem depender da goroutine agendada.
func (ar *Repository) InsertExpiredAuctionForTest(
	ctx context.Context, id string, pastTimestamp int64) error {
	return ar.InsertAuctionForTest(ctx, id,
		"test product", "test category", "test description for integration",
		auction.New, auction.Active, pastTimestamp)
}
