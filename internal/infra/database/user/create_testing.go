//go:build integration

package user

import (
	"context"
)

func (r *Repository) InsertUserForTest(ctx context.Context, id, name string) error {
	_, err := r.Collection.InsertOne(ctx, &document{ID: id, Name: name})
	return err
}
