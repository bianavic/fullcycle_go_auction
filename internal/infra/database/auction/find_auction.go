package auction

import (
	"context"
	"fmt"
	"fullcycle-auction_go/configuration/logger"
	"fullcycle-auction_go/internal/entity/auction"
	"fullcycle-auction_go/internal/internal_error"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (ar *AuctionRepository) FindAuctionByID(
	ctx context.Context, id string) (*auction.Auction, *internal_error.InternalError) {
	filter := bson.M{"_id": id}

	var auctionMongo AuctionMongo
	if err := ar.Collection.FindOne(ctx, filter).Decode(&auctionMongo); err != nil {
		logger.Error(fmt.Sprintf("Error trying to find auction by id = %s", id), err)
		return nil, internal_error.NewInternalServerError("Error trying to find auction by id")
	}

	return &auction.Auction{
		ID:          auctionMongo.ID,
		ProductName: auctionMongo.ProductName,
		Category:    auctionMongo.Category,
		Description: auctionMongo.Description,
		Condition:   auctionMongo.Condition,
		Status:      auctionMongo.Status,
		Timestamp:   time.Unix(auctionMongo.Timestamp, 0),
	}, nil
}

func (repo *AuctionRepository) FindAuctions(
	ctx context.Context,
	status auction.AuctionStatus,
	category string,
	productName string) ([]auction.Auction, *internal_error.InternalError) {
	filter := bson.M{}

	if status != 0 {
		filter["status"] = status
	}

	if category != "" {
		filter["category"] = category
	}

	if productName != "" {
		filter["product_name"] = primitive.Regex{Pattern: productName, Options: "i"}
	}

	cursor, err := repo.Collection.Find(ctx, filter)
	if err != nil {
		logger.Error("Error finding auctions", err)
		return nil, internal_error.NewInternalServerError("Error finding auctions")
	}
	defer func() { _ = cursor.Close(ctx) }()

	var auctionsMongo []AuctionMongo
	if err := cursor.All(ctx, &auctionsMongo); err != nil {
		logger.Error("Error decoding auctions", err)
		return nil, internal_error.NewInternalServerError("Error decoding auctions")
	}

	var auctions []auction.Auction
	for _, auctionMongo := range auctionsMongo {
		auctions = append(auctions, auction.Auction{
			ID:          auctionMongo.ID,
			ProductName: auctionMongo.ProductName,
			Category:    auctionMongo.Category,
			Status:      auctionMongo.Status,
			Description: auctionMongo.Description,
			Condition:   auctionMongo.Condition,
			Timestamp:   time.Unix(auctionMongo.Timestamp, 0),
		})
	}

	return auctions, nil
}
