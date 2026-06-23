package auction

import (
	"context"
	"fmt"
	"fullcycle-auction_go/configuration/logger"
	"fullcycle-auction_go/internal/apperr"
	"fullcycle-auction_go/internal/entity/auction"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (ar *Repository) FindByID(
	ctx context.Context, id string) (*auction.Auction, *apperr.InternalError) {
	filter := bson.M{"_id": id}

	var doc document
	if err := ar.Collection.FindOne(ctx, filter).Decode(&doc); err != nil {
		logger.Error(fmt.Sprintf("Error trying to find auction by id = %s", id), err)
		return nil, apperr.NewInternalServerError("Error trying to find auction by id")
	}

	return &auction.Auction{
		ID:          doc.ID,
		ProductName: doc.ProductName,
		Category:    doc.Category,
		Description: doc.Description,
		Condition:   doc.Condition,
		Status:      doc.Status,
		Timestamp:   time.Unix(doc.Timestamp, 0),
	}, nil
}

func (repo *Repository) FindAll(
	ctx context.Context,
	status auction.Status,
	category string,
	productName string) ([]auction.Auction, *apperr.InternalError) {
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
		return nil, apperr.NewInternalServerError("Error finding auctions")
	}
	defer func() { _ = cursor.Close(ctx) }()

	var docs []document
	if err := cursor.All(ctx, &docs); err != nil {
		logger.Error("Error decoding auctions", err)
		return nil, apperr.NewInternalServerError("Error decoding auctions")
	}

	var auctions []auction.Auction
	for _, doc := range docs {
		auctions = append(auctions, auction.Auction{
			ID:          doc.ID,
			ProductName: doc.ProductName,
			Category:    doc.Category,
			Status:      doc.Status,
			Description: doc.Description,
			Condition:   doc.Condition,
			Timestamp:   time.Unix(doc.Timestamp, 0),
		})
	}

	return auctions, nil
}
