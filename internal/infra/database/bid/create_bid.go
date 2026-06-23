package bid

import (
	"context"
	"fullcycle-auction_go/configuration/logger"
	"fullcycle-auction_go/internal/apperr"
	"fullcycle-auction_go/internal/config"
	"fullcycle-auction_go/internal/entity/auction"
	"fullcycle-auction_go/internal/entity/bid"
	auctionrepo "fullcycle-auction_go/internal/infra/database/auction"
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
		auctionInterval:       getAuctionInterval(),
		auctionStatusMap:      make(map[string]auction.Status),
		auctionEndTimeMap:     make(map[string]time.Time),
		auctionStatusMapMutex: &sync.Mutex{},
		auctionEndTimeMutex:   &sync.Mutex{},
		Collection:            database.Collection("bids"),
		AuctionRepository:     auctionRepository,
	}
}

func (bd *Repository) Create(
	ctx context.Context,
	bids []bid.Bid) *apperr.InternalError {
	var wg sync.WaitGroup
	for _, b := range bids {
		wg.Add(1)
		go func(bidValue bid.Bid) {
			defer wg.Done()

			bd.auctionStatusMapMutex.Lock()
			auctionStatus, okStatus := bd.auctionStatusMap[bidValue.AuctionID]
			bd.auctionStatusMapMutex.Unlock()

			bd.auctionEndTimeMutex.Lock()
			auctionEndTime, okEndTime := bd.auctionEndTimeMap[bidValue.AuctionID]
			bd.auctionEndTimeMutex.Unlock()

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

				if _, err := bd.Collection.InsertOne(ctx, doc); err != nil {
					logger.Error("Error trying to insert bid", err)
					return
				}

				return
			}

			found, err := bd.AuctionRepository.FindByID(ctx, bidValue.AuctionID)
			if err != nil {
				logger.Error("Error trying to find auction by id", err)
				return
			}
			if found.Status == auction.Completed {
				return
			}

			bd.auctionStatusMapMutex.Lock()
			bd.auctionStatusMap[bidValue.AuctionID] = found.Status
			bd.auctionStatusMapMutex.Unlock()

			bd.auctionEndTimeMutex.Lock()
			bd.auctionEndTimeMap[bidValue.AuctionID] = found.Timestamp.Add(bd.auctionInterval)
			bd.auctionEndTimeMutex.Unlock()

			if _, err := bd.Collection.InsertOne(ctx, doc); err != nil {
				logger.Error("Error trying to insert bid", err)
				return
			}
		}(b)
	}
	wg.Wait()
	return nil
}

func getAuctionInterval() time.Duration {
	return config.ParseDuration("AUCTION_INTERVAL", 5*time.Minute)
}
