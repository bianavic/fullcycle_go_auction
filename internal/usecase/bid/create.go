package bid

import (
	"context"
	"fullcycle-auction_go/internal/apperr"
	"fullcycle-auction_go/internal/entity/bid"
	"fullcycle-auction_go/internal/observability/logger"
	"os"
	"strconv"
	"time"
)

type InputDTO struct {
	UserID    string  `json:"user_id"`
	AuctionID string  `json:"auction_id"`
	Amount    float64 `json:"amount"`
}

type OutputDTO struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	AuctionID string    `json:"auction_id"`
	Amount    float64   `json:"amount"`
	Timestamp time.Time `json:"timestamp" time_format:"2006-01-02 15:04:05"`
}

type useCase struct {
	BidRepository bid.Repository

	timer               *time.Timer
	maxBatchSize        int
	batchInsertInterval time.Duration
	bidChannel          chan bid.Bid

	// batch acumula os bids pendentes. É acessado apenas pela goroutine de
	// triggerCreateRoutine, então não exige sincronização adicional.
	batch []bid.Bid
}

func New(ctx context.Context, bidRepository bid.Repository) UseCase {
	batchInsertInterval := getBatchInsertInterval()
	maxBatchSize := getMaxBatchSize()

	bidUseCase := &useCase{
		BidRepository:       bidRepository,
		maxBatchSize:        maxBatchSize,
		batchInsertInterval: batchInsertInterval,
		timer:               time.NewTimer(batchInsertInterval),
		bidChannel:          make(chan bid.Bid, maxBatchSize),
	}

	bidUseCase.triggerCreateRoutine(ctx)

	return bidUseCase
}

type UseCase interface {
	CreateBid(
		ctx context.Context,
		bidInputDTO InputDTO) *apperr.InternalError

	FindWinningBidByAuctionID(
		ctx context.Context, auctionID string) (*OutputDTO, *apperr.InternalError)

	FindBidByAuctionID(
		ctx context.Context, auctionID string) ([]OutputDTO, *apperr.InternalError)
}

func (uc *useCase) triggerCreateRoutine(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				// shutdown ordenado: drena os bids ainda no buffer e persiste o
				// lote acumulado antes de encerrar a goroutine.
				uc.drainAndFlush()
				return
			case newBid := <-uc.bidChannel:
				uc.batch = append(uc.batch, newBid)

				if len(uc.batch) >= uc.maxBatchSize {
					if err := uc.BidRepository.Create(ctx, uc.batch); err != nil {
						logger.Error("error trying to process bid batch list", err)
					}

					uc.batch = nil
					// Stop e drain antes do Reset: o timer pode ter expirado e deixado
					// um valor pendente em C, o que faria a próxima iteração entrar no
					// case do timer imediatamente. O guard len(uc.batch) > 0 cobre o
					// sintoma, mas o pattern correto é stop+drain+reset.
					if !uc.timer.Stop() {
						select {
						case <-uc.timer.C:
						default:
						}
					}
					uc.timer.Reset(uc.batchInsertInterval)
				}
			case <-uc.timer.C:
				if len(uc.batch) > 0 {
					if err := uc.BidRepository.Create(ctx, uc.batch); err != nil {
						logger.Error("error trying to process bid batch list", err)
					}
					uc.batch = nil
				}
				uc.timer.Reset(uc.batchInsertInterval)
			}
		}
	}()
}

// drainAndFlush esvazia o buffer do canal e persiste o lote acumulado. Usado no
// encerramento (ctx cancelado) para não perder bids já enfileirados. Usa um
// contexto próprio porque o contexto de ciclo de vida já foi cancelado.
func (uc *useCase) drainAndFlush() {
	for {
		select {
		case newBid := <-uc.bidChannel:
			uc.batch = append(uc.batch, newBid)
		default:
			if len(uc.batch) == 0 {
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			if err := uc.BidRepository.Create(ctx, uc.batch); err != nil {
				logger.Error("error flushing bid batch on shutdown", err)
			}
			cancel()
			uc.batch = nil
			return
		}
	}
}

func (uc *useCase) CreateBid(
	ctx context.Context,
	bidInputDTO InputDTO) *apperr.InternalError {

	newBid, err := bid.Create(bidInputDTO.UserID, bidInputDTO.AuctionID, bidInputDTO.Amount)
	if err != nil {
		return err
	}

	// Enfileira sem bloquear indefinidamente: se o contexto da requisição for
	// cancelado (cliente desconectou ou shutdown) enquanto o buffer está cheio,
	// retorna erro em vez de travar o handler HTTP.
	select {
	case uc.bidChannel <- *newBid:
		return nil
	case <-ctx.Done():
		logger.Error("context canceled before enqueueing bid", ctx.Err())
		return apperr.NewInternalServerError("could not enqueue bid")
	}
}

func getBatchInsertInterval() time.Duration {
	batchInsertInterval := os.Getenv("BATCH_INSERT_INTERVAL")
	duration, err := time.ParseDuration(batchInsertInterval)
	if err != nil {
		return 3 * time.Minute
	}

	return duration
}

func getMaxBatchSize() int {
	value, err := strconv.Atoi(os.Getenv("MAX_BATCH_SIZE"))
	if err != nil {
		return 5
	}

	return value
}
