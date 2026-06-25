package auction

import (
	"context"
	"fmt"
	"fullcycle-auction_go/internal/apperr"
	"fullcycle-auction_go/internal/config"
	"fullcycle-auction_go/internal/entity/auction"
	"fullcycle-auction_go/internal/observability/logger"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type document struct {
	ID          string                   `bson:"_id"`
	ProductName string                   `bson:"product_name"`
	Category    string                   `bson:"category"`
	Description string                   `bson:"description"`
	Condition   auction.ProductCondition `bson:"condition"`
	Status      auction.Status           `bson:"status"`
	Timestamp   int64                    `bson:"timestamp"`
}
type Repository struct {
	Collection      *mongo.Collection
	auctionInterval time.Duration
	auctionMutex    *sync.Mutex
	// closerCtx controla o ciclo de vida das goroutines de fechamento agendado.
	closerCtx context.Context
}

func New(ctx context.Context, database *mongo.Database) *Repository {
	return &Repository{
		Collection:      database.Collection("auctions"),
		auctionInterval: config.AuctionInterval(),
		auctionMutex:    &sync.Mutex{},
		closerCtx:       ctx,
	}
}

func (r *Repository) Create(
	ctx context.Context,
	auction *auction.Auction) *apperr.InternalError {
	doc := &document{
		ID:          auction.ID,
		ProductName: auction.ProductName,
		Category:    auction.Category,
		Description: auction.Description,
		Condition:   auction.Condition,
		Status:      auction.Status,
		Timestamp:   auction.Timestamp.Unix(),
	}
	_, err := r.Collection.InsertOne(ctx, doc)
	if err != nil {
		logger.Error("Error trying to insert auction", err)
		return apperr.NewInternalServerError("error trying to insert auction")
	}

	// Executa o fechamento de forma independente do ciclo de vida da requisição.
	go r.scheduleAuctionClose(doc.ID)

	return nil
}

// scheduleAuctionClose durante o shutdown, encerra sem fechar o leilão.
// Fechamentos perdidos são recuperados por StartAuctionCloser.
func (r *Repository) scheduleAuctionClose(auctionID string) {
	timer := time.NewTimer(r.auctionInterval)
	defer timer.Stop()

	select {
	case <-r.closerCtx.Done():
		return
	case <-timer.C:
	}

	if err := r.closeAuction(auctionID); err != nil {
		logger.Error(fmt.Sprintf("Error trying to close auction %s automatically", auctionID), err)
	}
}

// closeAuction usa um novo contexto porque o contexto da requisição pode
// já ter sido cancelado. O filtro por status mantém a operação idempotente.
func (r *Repository) closeAuction(auctionID string) *apperr.InternalError {
	r.auctionMutex.Lock()
	defer r.auctionMutex.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	filter := bson.M{"_id": auctionID, "status": auction.Active}
	update := bson.M{"$set": bson.M{"status": auction.Completed}}

	if _, err := r.Collection.UpdateOne(ctx, filter, update); err != nil {
		logger.Error("Error trying to close auction", err)
		return apperr.NewInternalServerError("error trying to close auction")
	}

	return nil
}

// StartAuctionCloser recupera leilões expirados que não foram fechados pelo
// agendamento, como após um reinício da aplicação.
func (r *Repository) StartAuctionCloser(ctx context.Context) {
	ticker := time.NewTicker(getCloserInterval())

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_ = r.closeExpiredAuctions(ctx)
			}
		}
	}()
}

// closeExpiredAuctions o filtro por status mantém a operação idempotente
// durante execuções concorrentes.
func (r *Repository) closeExpiredAuctions(ctx context.Context) *apperr.InternalError {
	r.auctionMutex.Lock()
	defer r.auctionMutex.Unlock()

	expirationLimit := time.Now().Add(-r.auctionInterval).Unix()

	filter := bson.M{
		"status":    auction.Active,
		"timestamp": bson.M{"$lt": expirationLimit},
	}
	update := bson.M{"$set": bson.M{"status": auction.Completed}}

	if _, err := r.Collection.UpdateMany(ctx, filter, update); err != nil {
		logger.Error("Error trying to close expired auctions", err)
		return apperr.NewInternalServerError("error trying to close expired auctions")
	}

	return nil
}

func getCloserInterval() time.Duration {
	return config.ParseDuration("AUCTION_CLOSER_INTERVAL", 10*time.Second)
}
