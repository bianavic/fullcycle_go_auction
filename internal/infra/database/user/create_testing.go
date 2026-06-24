//go:build integration

package user

import (
	"context"
)

func (ur *Repository) InsertUserForTest(ctx context.Context, id, name string) error {
	_, err := ur.Collection.InsertOne(ctx, &document{ID: id, Name: name})
	return err
}
