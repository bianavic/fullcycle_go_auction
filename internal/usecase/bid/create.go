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
				// Garante que bids pendentes sejam persistidos antes do shutdown.
				uc.drainAndFlush()
				return
			case newBid := <-uc.bidChannel:
				uc.batch = append(uc.batch, newBid)

				if len(uc.batch) >= uc.maxBatchSize {
					if err := uc.BidRepository.Create(ctx, uc.batch); err != nil {
						logger.Error("error trying to process bid batch list", err)
					}

					uc.batch = nil
					// Interrompe e drena o timer antes do Reset para
					// descartar um disparo pendente.
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

// drainAndFlush usa um novo contexto porque
// o contexto de ciclo de vida já foi cancelado.
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

	// Evita bloqueio indefinido quando a requisição é cancelada.
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
