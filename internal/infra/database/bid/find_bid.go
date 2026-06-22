package bid

import (
	"context"
	"fmt"
	"fullcycle-auction_go/configuration/logger"
	"fullcycle-auction_go/internal/entity/bid"
	"fullcycle-auction_go/internal/internal_error"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (bd *BidRepository) FindBidByAuctionID(
	ctx context.Context, auctionID string) ([]bid.Bid, *internal_error.InternalError) {
	filter := bson.M{"auction_id": auctionID}

	cursor, err := bd.Collection.Find(ctx, filter)
	if err != nil {
		logger.Error(
			fmt.Sprintf("Error trying to find bids by auctionID %s", auctionID), err)
		return nil, internal_error.NewInternalServerError(
			fmt.Sprintf("Error trying to find bids by auctionID %s", auctionID))
	}

	var bidsMongo []BidMongo
	if err := cursor.All(ctx, &bidsMongo); err != nil {
		logger.Error(
			fmt.Sprintf("Error trying to find bids by auctionID %s", auctionID), err)
		return nil, internal_error.NewInternalServerError(
			fmt.Sprintf("Error trying to find bids by auctionID %s", auctionID))
	}

	var bids []bid.Bid
	for _, bidMongo := range bidsMongo {
		bids = append(bids, bid.Bid{
			ID:        bidMongo.ID,
			UserID:    bidMongo.UserID,
			AuctionID: bidMongo.AuctionID,
			Amount:    bidMongo.Amount,
			Timestamp: time.Unix(bidMongo.Timestamp, 0),
		})
	}

	return bids, nil
}

func (bd *BidRepository) FindWinningBidByAuctionID(
	ctx context.Context, auctionID string) (*bid.Bid, *internal_error.InternalError) {
	filter := bson.M{"auction_id": auctionID}

	var bidMongo BidMongo
	opts := options.FindOne().SetSort(bson.D{{Key: "amount", Value: -1}})
	if err := bd.Collection.FindOne(ctx, filter, opts).Decode(&bidMongo); err != nil {
		logger.Error("Error trying to find the auction winner", err)
		return nil, internal_error.NewInternalServerError("Error trying to find the auction winner")
	}

	return &bid.Bid{
		ID:        bidMongo.ID,
		UserID:    bidMongo.UserID,
		AuctionID: bidMongo.AuctionID,
		Amount:    bidMongo.Amount,
		Timestamp: time.Unix(bidMongo.Timestamp, 0),
	}, nil
}
