//go:build integration

package auction_test

import (
	"context"
	"testing"
	"time"

	"fullcycle-auction_go/internal/entity/auction_entity"
	"fullcycle-auction_go/internal/infra/database/auction"

	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// setupMongo sobe um MongoDB efêmero via Testcontainers e devolve um *mongo.Database
// pronto para uso. A limpeza (encerrar container e desconectar o client) é registrada
// com t.Cleanup.
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

// waitForStatus consulta o leilão repetidamente até que ele atinja o status
// esperado ou o timeout estoure.
func waitForStatus(
	t *testing.T,
	repo *auction.AuctionRepository,
	id string,
	want auction_entity.AuctionStatus,
	timeout time.Duration,
) {
	t.Helper()
	ctx := context.Background()
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		found, err := repo.FindAuctionById(ctx, id)
		if err == nil && found.Status == want {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}

	t.Fatalf("auction %s did not reach status %d within %s", id, want, timeout)
}

// TestCreateAuction_ClosesAutomaticallyAfterInterval valida o fechamento agendado:
// ao criar um leilão, ele deve nascer Active e ser marcado como Completed assim que
// AUCTION_INTERVAL expira.
func TestCreateAuction_ClosesAutomaticallyAfterInterval(t *testing.T) {
	t.Setenv("AUCTION_INTERVAL", "1s")

	db := setupMongo(t)
	repo := auction.NewAuctionRepository(db)

	ctx := context.Background()

	auctionEntity, errEntity := auction_entity.CreateAuction(
		"Vintage Clock", "Decor", "A beautiful vintage wall clock", auction_entity.New)
	if errEntity != nil {
		t.Fatalf("failed to build auction entity: %v", errEntity.Message)
	}

	if err := repo.CreateAuction(ctx, auctionEntity); err != nil {
		t.Fatalf("failed to create auction: %v", err.Message)
	}

	found, err := repo.FindAuctionById(ctx, auctionEntity.Id)
	if err != nil {
		t.Fatalf("failed to find auction: %v", err.Message)
	}
	if found.Status != auction_entity.Active {
		t.Fatalf("expected auction to start as Active, got status %d", found.Status)
	}

	waitForStatus(t, repo, auctionEntity.Id, auction_entity.Completed, 5*time.Second)
}

// TestStartAuctionCloser_ClosesExpiredAuction valida o monitor em background:
// um leilão já vencido inserido diretamente no banco (sem passar pelo fechamento
// agendado) deve ser fechado pela varredura periódica.
func TestStartAuctionCloser_ClosesExpiredAuction(t *testing.T) {
	t.Setenv("AUCTION_INTERVAL", "1s")
	t.Setenv("AUCTION_CLOSER_INTERVAL", "1s")

	db := setupMongo(t)
	repo := auction.NewAuctionRepository(db)

	insertCtx := context.Background()
	expired := auction.AuctionEntityMongo{
		Id:          uuid.NewString(),
		ProductName: "Old Painting",
		Category:    "Art",
		Description: "An expired auction inserted directly into the database",
		Condition:   auction_entity.New,
		Status:      auction_entity.Active,
		Timestamp:   time.Now().Add(-time.Hour).Unix(),
	}
	if _, err := db.Collection("auctions").InsertOne(insertCtx, expired); err != nil {
		t.Fatalf("failed to insert expired auction: %v", err)
	}

	monitorCtx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	repo.StartAuctionCloser(monitorCtx)

	waitForStatus(t, repo, expired.Id, auction_entity.Completed, 5*time.Second)
}
