package user

import (
	"context"
	"errors"
	"fmt"
	"fullcycle-auction_go/configuration/logger"
	"fullcycle-auction_go/internal/entity/user"
	"fullcycle-auction_go/internal/internal_error"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserMongo struct {
	ID   string `bson:"_id"`
	Name string `bson:"name"`
}

type UserRepository struct {
	Collection *mongo.Collection
}

func New(database *mongo.Database) *UserRepository {
	return &UserRepository{
		Collection: database.Collection("users"),
	}
}

func (ur *UserRepository) FindUserByID(
	ctx context.Context, userID string) (*user.User, *internal_error.InternalError) {
	filter := bson.M{"_id": userID}

	var userMongo UserMongo
	err := ur.Collection.FindOne(ctx, filter).Decode(&userMongo)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			logger.Error(fmt.Sprintf("User not found with this id = %s", userID), err)
			return nil, internal_error.NewNotFoundError(
				fmt.Sprintf("User not found with this id = %s", userID))
		}

		logger.Error("Error trying to find user by userID", err)
		return nil, internal_error.NewInternalServerError("Error trying to find user by userID")
	}

	user := &user.User{
		ID:   userMongo.ID,
		Name: userMongo.Name,
	}

	return user, nil
}
