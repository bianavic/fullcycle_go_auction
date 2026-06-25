package bid

import (
	"context"
	"fullcycle-auction_go/internal/apperr"
	"fullcycle-auction_go/internal/config"
	"fullcycle-auction_go/internal/entity/auction"
	"fullcycle-auction_go/internal/entity/bid"
	auctionrepo "fullcycle-auction_go/internal/infra/database/auction"
	"fullcycle-auction_go/internal/observability/logger"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

type document struct {
	ID        string  `bson:"_id"`
	UserID    string  `bson:"user_id"`
	AuctionID string  `bson:"auction_id"`
	Amount    float64 `bson:"amount"`
	Timestamp int64   `bson:"timestamp"`
}

type Repository struct {
	Collection            *mongo.Collection
	AuctionRepository     *auctionrepo.Repository
	auctionInterval       time.Duration
	auctionStatusMap      map[string]auction.Status
	auctionEndTimeMap     map[string]time.Time
	auctionStatusMapMutex *sync.Mutex
	auctionEndTimeMutex   *sync.Mutex
}

func New(database *mongo.Database, auctionRepository *auctionrepo.Repository) *Repository {
	return &Repository{
		auctionInterval:       config.AuctionInterval(),
		auctionStatusMap:      make(map[string]auction.Status),
		auctionEndTimeMap:     make(map[string]time.Time),
		auctionStatusMapMutex: &sync.Mutex{},
		auctionEndTimeMutex:   &sync.Mutex{},
		Collection:            database.Collection("bids"),
		AuctionRepository:     auctionRepository,
	}
}

func (r *Repository) Create(
	ctx context.Context,
	bids []bid.Bid) *apperr.InternalError {
	var wg sync.WaitGroup
	for _, b := range bids {
		wg.Add(1)
		go func(bidValue bid.Bid) {
			defer wg.Done()

			r.auctionStatusMapMutex.Lock()
			auctionStatus, okStatus := r.auctionStatusMap[bidValue.AuctionID]
			r.auctionStatusMapMutex.Unlock()

			r.auctionEndTimeMutex.Lock()
			auctionEndTime, okEndTime := r.auctionEndTimeMap[bidValue.AuctionID]
			r.auctionEndTimeMutex.Unlock()

			doc := &document{
				ID:        bidValue.ID,
				UserID:    bidValue.UserID,
				AuctionID: bidValue.AuctionID,
				Amount:    bidValue.Amount,
				Timestamp: bidValue.Timestamp.Unix(),
			}

			if okEndTime && okStatus {
				now := time.Now()
				if auctionStatus == auction.Completed || now.After(auctionEndTime) {
					return
				}

				if _, err := r.Collection.InsertOne(ctx, doc); err != nil {
					logger.Error("Error trying to insert bid", err)
					return
				}

				return
			}

			found, err := r.AuctionRepository.FindByID(ctx, bidValue.AuctionID)
			if err != nil {
				logger.Error("Error trying to find auction by id", err)
				return
			}
			if found.Status == auction.Completed {
				return
			}

			r.auctionStatusMapMutex.Lock()
			r.auctionStatusMap[bidValue.AuctionID] = found.Status
			r.auctionStatusMapMutex.Unlock()

			r.auctionEndTimeMutex.Lock()
			r.auctionEndTimeMap[bidValue.AuctionID] = found.Timestamp.Add(r.auctionInterval)
			r.auctionEndTimeMutex.Unlock()

			if _, err := r.Collection.InsertOne(ctx, doc); err != nil {
				logger.Error("Error trying to insert bid", err)
				return
			}
		}(b)
	}
	wg.Wait()
	return nil
}
