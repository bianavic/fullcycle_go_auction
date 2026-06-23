package user

import (
	"context"
	"errors"
	"fmt"
	"fullcycle-auction_go/configuration/logger"
	"fullcycle-auction_go/internal/apperr"
	"fullcycle-auction_go/internal/entity/user"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type document struct {
	ID   string `bson:"_id"`
	Name string `bson:"name"`
}

type Repository struct {
	Collection *mongo.Collection
}

func New(database *mongo.Database) *Repository {
	return &Repository{
		Collection: database.Collection("users"),
	}
}

func (ur *Repository) FindByID(
	ctx context.Context, userID string) (*user.User, *apperr.InternalError) {
	filter := bson.M{"_id": userID}

	var doc document
	err := ur.Collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			logger.Error(fmt.Sprintf("User not found with this id = %s", userID), err)
			return nil, apperr.NewNotFoundError(
				fmt.Sprintf("User not found with this id = %s", userID))
		}

		logger.Error("Error trying to find user by userID", err)
		return nil, apperr.NewInternalServerError("Error trying to find user by userID")
	}

	user := &user.User{
		ID:   doc.ID,
		Name: doc.Name,
	}

	return user, nil
}
