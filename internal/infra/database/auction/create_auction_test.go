//go:build integration

package auction_test

import (
	"context"
	"os"
	"testing"
	"time"

	"fullcycle-auction_go/internal/entity/auction_entity"
	"fullcycle-auction_go/internal/infra/database/auction"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TestMain define intervalos curtos como padrão do pacote para que os testes de
// fechamento automático rodem rápido. As variáveis são definidas aqui (e não via
// t.Setenv) porque t.Setenv é incompatível com t.Parallel — definindo o ambiente
// uma única vez, antes de qualquer teste, os testes podem rodar em paralelo com
// segurança (leitura somente durante a execução).
func TestMain(m *testing.M) {
	_ = os.Setenv("AUCTION_INTERVAL", "1s")
	_ = os.Setenv("AUCTION_CLOSER_INTERVAL", "1s")
	os.Exit(m.Run())
}

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

	require.Eventually(t, func() bool {
		found, err := repo.FindAuctionById(ctx, id)
		return err == nil && found.Status == want
	}, timeout, 100*time.Millisecond,
		"auction %s did not reach status %d within %s", id, want, timeout)
}

// TestCreateAuction_ClosesAutomaticallyAfterInterval valida o fechamento agendado:
// ao criar um leilão, ele deve nascer Active e ser marcado como Completed assim que
// AUCTION_INTERVAL expira.
func TestCreateAuction_ClosesAutomaticallyAfterInterval(t *testing.T) {
	t.Parallel()

	db := setupMongo(t)
	repo := auction.NewAuctionRepository(db)
	ctx := context.Background()

	auctionEntity, errEntity := auction_entity.CreateAuction(
		"Vintage Clock", "Decor", "A beautiful vintage wall clock", auction_entity.New)
	require.Nil(t, errEntity, "failed to build auction entity")

	require.Nil(t, repo.CreateAuction(ctx, auctionEntity), "failed to create auction")

	t.Run("initial status is Active", func(t *testing.T) {
		found, err := repo.FindAuctionById(ctx, auctionEntity.Id)
		require.Nil(t, err, "failed to find auction")
		require.Equal(t, auction_entity.Active, found.Status)
	})

	t.Run("status becomes Completed after interval", func(t *testing.T) {
		waitForStatus(t, repo, auctionEntity.Id, auction_entity.Completed, 5*time.Second)
	})
}

// TestStartAuctionCloser_ClosesExpiredAuction valida o monitor em background:
// um leilão já vencido inserido diretamente no banco (sem passar pelo fechamento
// agendado) deve ser fechado pela varredura periódica.
func TestStartAuctionCloser_ClosesExpiredAuction(t *testing.T) {
	t.Parallel()

	db := setupMongo(t)
	repo := auction.NewAuctionRepository(db)
	ctx := context.Background()

	id := uuid.NewString()
	require.NoError(t, repo.InsertExpiredAuctionForTest(ctx, id, time.Now().Add(-time.Hour).Unix()))

	monitorCtx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	repo.StartAuctionCloser(monitorCtx)

	waitForStatus(t, repo, id, auction_entity.Completed, 5*time.Second)
}

// TestStartAuctionCloser_StopsOnContextCancel valida o ciclo de vida da goroutine
// do monitor: após o cancelamento do contexto, a varredura para e um leilão vencido
// inserido em seguida permanece Active.
func TestStartAuctionCloser_StopsOnContextCancel(t *testing.T) {
	t.Parallel()

	db := setupMongo(t)
	repo := auction.NewAuctionRepository(db)
	ctx := context.Background()

	// confirma que o monitor está vivo fechando um primeiro leilão vencido;
	// ao final do Eventually a goroutine está ociosa (bloqueada no select).
	aliveID := uuid.NewString()
	require.NoError(t, repo.InsertExpiredAuctionForTest(ctx, aliveID, time.Now().Add(-time.Hour).Unix()))

	monitorCtx, cancel := context.WithCancel(context.Background())
	repo.StartAuctionCloser(monitorCtx)
	waitForStatus(t, repo, aliveID, auction_entity.Completed, 5*time.Second)

	// com o monitor ocioso, o cancelamento é observado imediatamente no select e a
	// goroutine encerra sem nova varredura.
	cancel()

	stoppedID := uuid.NewString()
	require.NoError(t, repo.InsertExpiredAuctionForTest(ctx, stoppedID, time.Now().Add(-time.Hour).Unix()))

	require.Never(t, func() bool {
		found, err := repo.FindAuctionById(ctx, stoppedID)
		return err == nil && found.Status == auction_entity.Completed
	}, 3*time.Second, 200*time.Millisecond,
		"auction %s should remain Active after the closer context was cancelled", stoppedID)
}
