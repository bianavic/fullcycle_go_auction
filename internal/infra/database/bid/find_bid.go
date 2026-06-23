package bid

import (
	"context"
	"fmt"
	"fullcycle-auction_go/configuration/logger"
	"fullcycle-auction_go/internal/apperr"
	"fullcycle-auction_go/internal/entity/bid"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (bd *Repository) FindByAuctionID(
	ctx context.Context, auctionID string) ([]bid.Bid, *apperr.InternalError) {
	filter := bson.M{"auction_id": auctionID}

	cursor, err := bd.Collection.Find(ctx, filter)
	if err != nil {
		logger.Error(
			fmt.Sprintf("Error trying to find bids by auctionID %s", auctionID), err)
		return nil, apperr.NewInternalServerError(
			fmt.Sprintf("Error trying to find bids by auctionID %s", auctionID))
	}

	var docs []document
	if err := cursor.All(ctx, &docs); err != nil {
		logger.Error(
			fmt.Sprintf("Error trying to find bids by auctionID %s", auctionID), err)
		return nil, apperr.NewInternalServerError(
			fmt.Sprintf("Error trying to find bids by auctionID %s", auctionID))
	}

	var bids []bid.Bid
	for _, doc := range docs {
		bids = append(bids, bid.Bid{
			ID:        doc.ID,
			UserID:    doc.UserID,
			AuctionID: doc.AuctionID,
			Amount:    doc.Amount,
			Timestamp: time.Unix(doc.Timestamp, 0),
		})
	}

	return bids, nil
}

func (bd *Repository) FindWinningByAuctionID(
	ctx context.Context, auctionID string) (*bid.Bid, *apperr.InternalError) {
	filter := bson.M{"auction_id": auctionID}

	var doc document
	opts := options.FindOne().SetSort(bson.D{{Key: "amount", Value: -1}})
	if err := bd.Collection.FindOne(ctx, filter, opts).Decode(&doc); err != nil {
		logger.Error("Error trying to find the auction winner", err)
		return nil, apperr.NewInternalServerError("Error trying to find the auction winner")
	}

	return &bid.Bid{
		ID:        doc.ID,
		UserID:    doc.UserID,
		AuctionID: doc.AuctionID,
		Amount:    doc.Amount,
		Timestamp: time.Unix(doc.Timestamp, 0),
	}, nil
}
