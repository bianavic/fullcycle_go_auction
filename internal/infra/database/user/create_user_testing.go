//go:build integration

package user

import (
	"context"
)

// InsertUserForTest insere um user diretamente na coleção. Usado por testes de
// integração que precisam de um user pré-existente para exercitar a busca.
func (ur *UserRepository) InsertUserForTest(ctx context.Context, id, name string) error {
	_, err := ur.Collection.InsertOne(ctx, &UserMongo{ID: id, Name: name})
	return err
}
