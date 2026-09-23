package main

import (
	"context"
	"errors"
	"fullcycle-auction_go/internal/app"
	"fullcycle-auction_go/internal/infra/api/web"
	"fullcycle-auction_go/internal/infra/database/mongodb"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
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

	userController, bidController, auctionsController := app.BuildDependencies(ctx, databaseConnection)
	router := web.NewRouter(userController, bidController, auctionsController)

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
