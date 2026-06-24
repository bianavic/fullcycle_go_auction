package main

import (
	"context"
	"errors"
	auctioncontroller "fullcycle-auction_go/internal/infra/api/web/controller/auction"
	"fullcycle-auction_go/internal/infra/api/web/controller/bid"
	"fullcycle-auction_go/internal/infra/api/web/controller/user"
	auctionrepository "fullcycle-auction_go/internal/infra/database/auction"
	bidrepository "fullcycle-auction_go/internal/infra/database/bid"
	"fullcycle-auction_go/internal/infra/database/mongodb"
	userrepository "fullcycle-auction_go/internal/infra/database/user"
	auctionuc "fullcycle-auction_go/internal/usecase/auction"
	biduc "fullcycle-auction_go/internal/usecase/bid"
	useruc "fullcycle-auction_go/internal/usecase/user"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
)

func main() {
	// ctx é cancelado ao receber SIGINT/SIGTERM, propagando o shutdown para as
	// goroutines de background (fechamento de leilões e processamento de bids).
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

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

	server := &http.Server{Addr: ":8080", Handler: router}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server failed to start: %v", err)
		}
	}()

	<-ctx.Done()
	// restaura o comportamento padrão dos sinais: um segundo SIGINT/SIGTERM
	// passa a encerrar o processo imediatamente.
	stop()
	log.Println("shutdown signal received; draining connections")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("server forced to shutdown: %v", err)
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
		auctionuc.New(auctionRepository, bidRepository))
	bidController = bid.New(biduc.New(ctx, bidRepository))

	return
}
