package mongodb

import (
	"context"
	"fullcycle-auction_go/configuration/logger"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	mongoDBURL = "MONGODB_URL"
	mongoDBDB  = "MONGODB_DB"
)

func NewConnection(ctx context.Context) (*mongo.Database, error) {
	mongoURL := os.Getenv(mongoDBURL)
	mongoDatabase := os.Getenv(mongoDBDB)

	client, err := mongo.Connect(
		ctx, options.Client().ApplyURI(mongoURL))
	if err != nil {
		logger.Error("Error trying to connect to mongodb database", err)
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		logger.Error("Error trying to ping mongodb database", err)
		return nil, err
	}

	return client.Database(mongoDatabase), nil
}
