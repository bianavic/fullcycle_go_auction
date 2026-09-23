// Package app monta as dependências da aplicação. É importável tanto por
// cmd/auction/main.go quanto por testes E2E, que precisam do mesmo wiring
// de produção rodando contra um banco efêmero.
package app

import (
	"context"

	auctioncontroller "fullcycle-auction_go/internal/infra/api/web/controller/auction"
	"fullcycle-auction_go/internal/infra/api/web/controller/bid"
	"fullcycle-auction_go/internal/infra/api/web/controller/user"
	auctionrepository "fullcycle-auction_go/internal/infra/database/auction"
	bidrepository "fullcycle-auction_go/internal/infra/database/bid"
	userrepository "fullcycle-auction_go/internal/infra/database/user"
	auctionuc "fullcycle-auction_go/internal/usecase/auction"
	biduc "fullcycle-auction_go/internal/usecase/bid"
	useruc "fullcycle-auction_go/internal/usecase/user"

	"go.mongodb.org/mongo-driver/mongo"
)

// BuildDependencies monta os controllers e inicia as goroutines de background
// (fechamento automático de leilões e processamento de lances). O fechamento
// das goroutines é acionado pelo cancelamento de ctx.
func BuildDependencies(ctx context.Context, database *mongo.Database) (
	userController *user.Controller,
	bidController *bid.Controller,
	auctionController *auctioncontroller.Controller) {

	auctionRepository := auctionrepository.New(ctx, database)
	auctionRepository.StartAuctionCloser(ctx)

	bidRepository := bidrepository.New(database, auctionRepository)
	userRepository := userrepository.New(database)

	userController = user.New(
		useruc.New(userRepository))
	auctionController = auctioncontroller.New(
		auctionuc.New(auctionRepository, bidRepository))
	bidController = bid.New(biduc.New(ctx, bidRepository))

	return
}
