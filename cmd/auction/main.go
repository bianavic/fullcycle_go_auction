package main

import (
	"context"
	"fullcycle-auction_go/configuration/database/mongodb"
	auctioncontroller "fullcycle-auction_go/internal/infra/api/web/controller/auction"
	"fullcycle-auction_go/internal/infra/api/web/controller/bid"
	"fullcycle-auction_go/internal/infra/api/web/controller/user"
	auctionrepository "fullcycle-auction_go/internal/infra/database/auction"
	bidrepository "fullcycle-auction_go/internal/infra/database/bid"
	userrepository "fullcycle-auction_go/internal/infra/database/user"
	auctionuc "fullcycle-auction_go/internal/usecase/auction"
	biduc "fullcycle-auction_go/internal/usecase/bid"
	useruc "fullcycle-auction_go/internal/usecase/user"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
)

func main() {
	ctx := context.Background()

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found; relying on environment variables")
	}

	databaseConnection, err := mongodb.NewConnection(ctx)
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	router := gin.Default()

	userController, bidController, auctionsController := initDependencies(ctx, databaseConnection)

	router.GET("/auction", auctionsController.FindAuctions)
	router.GET("/auction/:auctionId", auctionsController.FindAuctionByID)
	router.POST("/auction", auctionsController.CreateAuction)
	router.GET("/auction/winner/:auctionId", auctionsController.FindWinningBidByAuctionID)
	router.POST("/bid", bidController.CreateBid)
	router.GET("/bid/:auctionId", bidController.FindBidByAuctionID)
	router.GET("/user/:userId", userController.FindUserByID)

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}

func initDependencies(ctx context.Context, database *mongo.Database) (
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
		ctx,
		auctionuc.New(auctionRepository, bidRepository))
	bidController = bid.New(biduc.New(bidRepository))

	return
}
