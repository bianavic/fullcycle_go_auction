package bid_usecase

import (
	"context"
	"fullcycle-auction_go/configuration/logger"
	"fullcycle-auction_go/internal/entity/bid_entity"
	"fullcycle-auction_go/internal/internal_error"
	"os"
	"strconv"
	"time"
)

type BidInputDTO struct {
	UserId    string  `json:"user_id"`
	AuctionId string  `json:"auction_id"`
	Amount    float64 `json:"amount"`
}

type BidOutputDTO struct {
	Id        string    `json:"id"`
	UserId    string    `json:"user_id"`
	AuctionId string    `json:"auction_id"`
	Amount    float64   `json:"amount"`
	Timestamp time.Time `json:"timestamp" time_format:"2006-01-02 15:04:05"`
}

type BidUseCase struct {
	BidRepository bid_entity.BidEntityRepository

	timer               *time.Timer
	maxBatchSize        int
	batchInsertInterval time.Duration
	bidChannel          chan bid_entity.Bid

	// batch acumula os bids pendentes. É acessado apenas pela goroutine de
	// triggerCreateRoutine, então não exige sincronização adicional.
	batch []bid_entity.Bid
}

func NewBidUseCase(bidRepository bid_entity.BidEntityRepository) BidUseCaseInterface {
	maxSizeInterval := getMaxBatchSizeInterval()
	maxBatchSize := getMaxBatchSize()

	bidUseCase := &BidUseCase{
		BidRepository:       bidRepository,
		maxBatchSize:        maxBatchSize,
		batchInsertInterval: maxSizeInterval,
		timer:               time.NewTimer(maxSizeInterval),
		bidChannel:          make(chan bid_entity.Bid, maxBatchSize),
	}

	bidUseCase.triggerCreateRoutine(context.Background())

	return bidUseCase
}

type BidUseCaseInterface interface {
	CreateBid(
		ctx context.Context,
		bidInputDTO BidInputDTO) *internal_error.InternalError

	FindWinningBidByAuctionId(
		ctx context.Context, auctionId string) (*BidOutputDTO, *internal_error.InternalError)

	FindBidByAuctionId(
		ctx context.Context, auctionId string) ([]BidOutputDTO, *internal_error.InternalError)
}

func (bu *BidUseCase) triggerCreateRoutine(ctx context.Context) {
	go func() {
		defer close(bu.bidChannel)

		for {
			select {
			case bidEntity, ok := <-bu.bidChannel:
				if !ok {
					if len(bu.batch) > 0 {
						if err := bu.BidRepository.CreateBid(ctx, bu.batch); err != nil {
							logger.Error("error trying to process bid batch list", err)
						}
					}
					return
				}

				bu.batch = append(bu.batch, bidEntity)

				if len(bu.batch) >= bu.maxBatchSize {
					if err := bu.BidRepository.CreateBid(ctx, bu.batch); err != nil {
						logger.Error("error trying to process bid batch list", err)
					}

					bu.batch = nil
					// Stop e drain antes do Reset: o timer pode ter expirado e deixado
					// um valor pendente em C, o que faria a próxima iteração entrar no
					// case do timer imediatamente. O guard len(bu.batch) > 0 cobre o
					// sintoma, mas o pattern correto é stop+drain+reset.
					if !bu.timer.Stop() {
						select {
						case <-bu.timer.C:
						default:
						}
					}
					bu.timer.Reset(bu.batchInsertInterval)
				}
			case <-bu.timer.C:
				if len(bu.batch) > 0 {
					if err := bu.BidRepository.CreateBid(ctx, bu.batch); err != nil {
						logger.Error("error trying to process bid batch list", err)
					}
					bu.batch = nil
				}
				bu.timer.Reset(bu.batchInsertInterval)
			}
		}
	}()
}

func (bu *BidUseCase) CreateBid(
	ctx context.Context,
	bidInputDTO BidInputDTO) *internal_error.InternalError {

	bidEntity, err := bid_entity.CreateBid(bidInputDTO.UserId, bidInputDTO.AuctionId, bidInputDTO.Amount)
	if err != nil {
		return err
	}

	bu.bidChannel <- *bidEntity

	return nil
}

func getMaxBatchSizeInterval() time.Duration {
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
