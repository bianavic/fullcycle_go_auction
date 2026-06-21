package auction

import (
	"context"
	"fmt"
	"fullcycle-auction_go/configuration/logger"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"fullcycle-auction_go/internal/internal_error"
	"os"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuctionEntityMongo struct {
	Id          string                          `bson:"_id"`
	ProductName string                          `bson:"product_name"`
	Category    string                          `bson:"category"`
	Description string                          `bson:"description"`
	Condition   auction_entity.ProductCondition `bson:"condition"`
	Status      auction_entity.AuctionStatus    `bson:"status"`
	Timestamp   int64                           `bson:"timestamp"`
}
type AuctionRepository struct {
	Collection      *mongo.Collection
	auctionInterval time.Duration
	auctionMutex    *sync.Mutex
	// closerCtx limita o tempo de vida das goroutines de fechamento agendado.
	// Quando cancelado, scheduleAuctionClose retorna sem disparar o update.
	closerCtx context.Context
}

func NewAuctionRepository(ctx context.Context, database *mongo.Database) *AuctionRepository {
	return &AuctionRepository{
		Collection:      database.Collection("auctions"),
		auctionInterval: getAuctionInterval(),
		auctionMutex:    &sync.Mutex{},
		closerCtx:       ctx,
	}
}

func (ar *AuctionRepository) CreateAuction(
	ctx context.Context,
	auctionEntity *auction_entity.Auction) *internal_error.InternalError {
	auctionEntityMongo := &AuctionEntityMongo{
		Id:          auctionEntity.Id,
		ProductName: auctionEntity.ProductName,
		Category:    auctionEntity.Category,
		Description: auctionEntity.Description,
		Condition:   auctionEntity.Condition,
		Status:      auctionEntity.Status,
		Timestamp:   auctionEntity.Timestamp.Unix(),
	}
	_, err := ar.Collection.InsertOne(ctx, auctionEntityMongo)
	if err != nil {
		logger.Error("Error trying to insert auction", err)
		return internal_error.NewInternalServerError("Error trying to insert auction")
	}

	// agenda o fechamento pontual deste leilão após o intervalo configurado.
	// A goroutine sobrevive ao ciclo de vida do request HTTP.
	go ar.scheduleAuctionClose(auctionEntityMongo.Id)

	return nil
}

// scheduleAuctionClose aguarda o intervalo do leilão e, ao expirar, dispara o
// fechamento. Roda em uma goroutine independente do request que criou o leilão.
// O timer respeita closerCtx para permitir shutdown ordenado: se o contexto
// for cancelado antes do intervalo, a goroutine retorna sem fechar o leilão
// (a varredura do StartAuctionCloser cobre o caso após restart).
func (ar *AuctionRepository) scheduleAuctionClose(auctionId string) {
	timer := time.NewTimer(ar.auctionInterval)
	defer timer.Stop()

	select {
	case <-ar.closerCtx.Done():
		return
	case <-timer.C:
	}

	if err := ar.closeAuction(auctionId); err != nil {
		logger.Error(fmt.Sprintf("Error trying to close auction %s automatically", auctionId), err)
	}
}

// closeAuction atualiza o status do leilão para Completed. O filtro por
// status=Active garante idempotência (não reabre nem reescreve um leilão já
// fechado) e evita corridas entre a goroutine pontual e o monitor em background.
// Usa um contexto próprio, pois o contexto do request original já pode ter sido
// cancelado quando o fechamento ocorre.
func (ar *AuctionRepository) closeAuction(auctionId string) *internal_error.InternalError {
	ar.auctionMutex.Lock()
	defer ar.auctionMutex.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	filter := bson.M{"_id": auctionId, "status": auction_entity.Active}
	update := bson.M{"$set": bson.M{"status": auction_entity.Completed}}

	if _, err := ar.Collection.UpdateOne(ctx, filter, update); err != nil {
		logger.Error("Error trying to close auction", err)
		return internal_error.NewInternalServerError("Error trying to close auction")
	}

	return nil
}

// StartAuctionCloser inicia um monitor que varre o banco periodicamente em busca
// de leilões vencidos ainda marcados como Active e os fecha em lote. Funciona como
// rede de segurança para leilões cujo fechamento agendado foi perdido (por exemplo,
// após um restart do processo). O monitor encerra quando o contexto é cancelado.
func (ar *AuctionRepository) StartAuctionCloser(ctx context.Context) {
	ticker := time.NewTicker(getCloserInterval())

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_ = ar.closeExpiredAuctions(ctx)
			}
		}
	}()
}

// closeExpiredAuctions fecha todos os leilões Active cujo tempo já expirou
// (timestamp + auctionInterval < agora). O filtro por status=Active mantém a
// operação idempotente e segura quando concorre com o fechamento agendado.
func (ar *AuctionRepository) closeExpiredAuctions(ctx context.Context) *internal_error.InternalError {
	ar.auctionMutex.Lock()
	defer ar.auctionMutex.Unlock()

	expirationLimit := time.Now().Add(-ar.auctionInterval).Unix()

	filter := bson.M{
		"status":    auction_entity.Active,
		"timestamp": bson.M{"$lt": expirationLimit},
	}
	update := bson.M{"$set": bson.M{"status": auction_entity.Completed}}

	if _, err := ar.Collection.UpdateMany(ctx, filter, update); err != nil {
		logger.Error("Error trying to close expired auctions", err)
		return internal_error.NewInternalServerError("Error trying to close expired auctions")
	}

	return nil
}

// getAuctionInterval calcula a duração do leilão a partir da variável de
// ambiente AUCTION_INTERVAL (ex.: "20s", "5m"). Caso a variável esteja ausente
// ou seja inválida, assume 5 minutos como padrão.
func getAuctionInterval() time.Duration {
	auctionInterval := os.Getenv("AUCTION_INTERVAL")
	duration, err := time.ParseDuration(auctionInterval)
	if err != nil {
		return time.Minute * 5
	}

	return duration
}

// getCloserInterval define a frequência de varredura do monitor de fechamento,
// a partir de AUCTION_CLOSER_INTERVAL. Default de 10 segundos quando a variável
// está ausente ou é inválida.
func getCloserInterval() time.Duration {
	closerInterval := os.Getenv("AUCTION_CLOSER_INTERVAL")
	duration, err := time.ParseDuration(closerInterval)
	if err != nil {
		return time.Second * 10
	}

	return duration
}
