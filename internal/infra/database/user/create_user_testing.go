//go:build integration

package user

import (
	"context"
)

func (ur *UserRepository) InsertUserForTest(ctx context.Context, id, name string) error {
	_, err := ur.Collection.InsertOne(ctx, &UserMongo{ID: id, Name: name})
	return err
}
