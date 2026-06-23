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
	// closerCtx limita o tempo de vida das goroutines de fechamento agendado.
	// Quando cancelado, scheduleAuctionClose retorna sem disparar o update.
	closerCtx context.Context
}

func New(ctx context.Context, database *mongo.Database) *Repository {
	return &Repository{
		Collection:      database.Collection("auctions"),
		auctionInterval: getAuctionInterval(),
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
		return apperr.NewInternalServerError("Error trying to insert auction")
	}

	// agenda o fechamento pontual deste leilão após o intervalo configurado.
	// A goroutine sobrevive ao ciclo de vida do request HTTP.
	go r.scheduleAuctionClose(doc.ID)

	return nil
}

// scheduleAuctionClose aguarda o intervalo do leilão e, ao expirar, dispara o
// fechamento. Roda em uma goroutine independente do request que criou o leilão.
// O timer respeita closerCtx para permitir shutdown ordenado: se o contexto
// for cancelado antes do intervalo, a goroutine retorna sem fechar o leilão
// (a varredura do StartAuctionCloser cobre o caso após restart).
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

// closeAuction atualiza o status do leilão para Completed. O filtro por
// status=Active garante idempotência (não reabre nem reescreve um leilão já
// fechado) e evita corridas entre a goroutine pontual e o monitor em background.
// Usa um contexto próprio, pois o contexto do request original já pode ter sido
// cancelado quando o fechamento ocorre.
func (r *Repository) closeAuction(auctionID string) *apperr.InternalError {
	r.auctionMutex.Lock()
	defer r.auctionMutex.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	filter := bson.M{"_id": auctionID, "status": auction.Active}
	update := bson.M{"$set": bson.M{"status": auction.Completed}}

	if _, err := r.Collection.UpdateOne(ctx, filter, update); err != nil {
		logger.Error("Error trying to close auction", err)
		return apperr.NewInternalServerError("Error trying to close auction")
	}

	return nil
}

// StartAuctionCloser inicia um monitor que varre o banco periodicamente em busca
// de leilões vencidos ainda marcados como Active e os fecha em lote. Funciona como
// rede de segurança para leilões cujo fechamento agendado foi perdido (por exemplo,
// após um restart do processo). O monitor encerra quando o contexto é cancelado.
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

// closeExpiredAuctions fecha todos os leilões Active cujo tempo já expirou
// (timestamp + auctionInterval < agora). O filtro por status=Active mantém a
// operação idempotente e segura quando concorre com o fechamento agendado.
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
		return apperr.NewInternalServerError("Error trying to close expired auctions")
	}

	return nil
}

func getAuctionInterval() time.Duration {
	return config.ParseDuration("AUCTION_INTERVAL", 5*time.Minute)
}

func getCloserInterval() time.Duration {
	return config.ParseDuration("AUCTION_CLOSER_INTERVAL", 10*time.Second)
}
