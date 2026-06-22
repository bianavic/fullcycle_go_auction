//go:build integration

package user_test

import (
	"context"
	"testing"

	"fullcycle-auction_go/internal/infra/database/user"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func setupMongo(t *testing.T) *mongo.Database {
	t.Helper()
	ctx := context.Background()

	container, err := mongodb.Run(ctx, "mongo:7")
	if err != nil {
		t.Fatalf("failed to start mongodb container: %v", err)
	}
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			t.Logf("failed to terminate mongodb container: %v", err)
		}
	})

	uri, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatalf("failed to connect to mongodb: %v", err)
	}
	t.Cleanup(func() {
		_ = client.Disconnect(ctx)
	})

	if err := client.Ping(ctx, nil); err != nil {
		t.Fatalf("failed to ping mongodb: %v", err)
	}

	return client.Database("auctions_test")
}

func TestFindUserByID(t *testing.T) {
	t.Parallel()

	t.Run("found", func(t *testing.T) {
		t.Parallel()
		db := setupMongo(t)
		repo := user.New(db)
		ctx := context.Background()

		id := uuid.NewString()
		require.NoError(t, repo.InsertUserForTest(ctx, id, "Alice"))

		found, err := repo.FindUserByID(ctx, id)
		require.Nil(t, err)
		require.Equal(t, id, found.ID)
		require.Equal(t, "Alice", found.Name)
	})

	// O marcador %! aparece quando há erro de verbo de formato. exemplo: %d -> %s
	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		db := setupMongo(t)
		repo := user.New(db)
		ctx := context.Background()

		id := uuid.NewString()
		found, err := repo.FindUserByID(ctx, id)
		require.NotNil(t, err)
		require.Nil(t, found)
		require.Equal(t, "not_found", err.Err)
		require.Contains(t, err.Message, id)
		require.NotContains(t, err.Message, "%!")
	})
}