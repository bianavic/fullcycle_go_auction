//go:build integration

package auction_test

import (
	"context"
	"os"
	"testing"
	"time"

	"fullcycle-auction_go/internal/entity/auction"
	auctiondb "fullcycle-auction_go/internal/infra/database/auction"

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

func waitForStatus(
	t *testing.T,
	repo *auctiondb.Repository,
	id string,
	want auction.Status,
	timeout time.Duration,
) {
	t.Helper()
	ctx := context.Background()

	require.Eventually(t, func() bool {
		found, err := repo.FindByID(ctx, id)
		return err == nil && found.Status == want
	}, timeout, 100*time.Millisecond,
		"auction %s did not reach status %d within %s", id, want, timeout)
}

func TestCreateAuction_ClosesAutomaticallyAfterInterval(t *testing.T) {
	t.Parallel()

	db := setupMongo(t)
	ctx := context.Background()
	repo := auctiondb.New(ctx, db)

	auctionEntity, errEntity := auction.Create(
		"Vintage Clock", "Decor", "A beautiful vintage wall clock", auction.New)
	require.Nil(t, errEntity, "failed to build auction entity")

	require.Nil(t, repo.Create(ctx, auctionEntity), "failed to create auction")

	t.Run("initial status is Active", func(t *testing.T) {
		found, err := repo.FindByID(ctx, auctionEntity.ID)
		require.Nil(t, err, "failed to find auction")
		require.Equal(t, auction.Active, found.Status)
	})

	t.Run("status becomes Completed after interval", func(t *testing.T) {
		waitForStatus(t, repo, auctionEntity.ID, auction.Completed, 5*time.Second)
	})
}

func TestStartAuctionCloser(t *testing.T) {
	t.Parallel()

	// Valida o monitor em background: um leilão já vencido inserido diretamente no banco
	// (sem passar pelo fechamento agendado) deve ser fechado pela varredura periódica.
	t.Run("closes expired auction", func(t *testing.T) {
		t.Parallel()
		db := setupMongo(t)
		ctx := context.Background()
		repo := auctiondb.New(ctx, db)

		id := uuid.NewString()
		require.NoError(t, repo.InsertExpiredAuctionForTest(ctx, id, time.Now().Add(-time.Hour).Unix()))

		monitorCtx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)
		repo.StartAuctionCloser(monitorCtx)

		waitForStatus(t, repo, id, auction.Completed, 5*time.Second)
	})

	// Verifica que o monitor nunca reabre um leilão já fechado: o filtro status=Active em
	// closeExpiredAuctions impede qualquer atualização em documentos com status=Completed,
	// mesmo quando o timestamp do leilão está dentro da janela de vencimento.
	t.Run("completed auction is not reopened", func(t *testing.T) {
		t.Parallel()
		db := setupMongo(t)
		ctx := context.Background()
		repo := auctiondb.New(ctx, db)

		id := uuid.NewString()
		require.NoError(t, repo.InsertAuctionForTest(ctx, id,
			"Completed Item", "Cat", "a completed auction for integration",
			auction.New, auction.Completed, time.Now().Add(-time.Hour).Unix()))

		monitorCtx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)
		repo.StartAuctionCloser(monitorCtx)

		require.Never(t, func() bool {
			found, err := repo.FindByID(ctx, id)
			return err != nil || found.Status != auction.Completed
		}, 3*time.Second, 200*time.Millisecond,
			"auction %s was modified after already being Completed", id)
	})

	// Verifica que o monitor não fecha leilões Active cujo timestamp está no futuro:
	// o filtro $lt não os alcança e o status permanece Active após vários ticks.
	t.Run("no expired auctions does nothing", func(t *testing.T) {
		t.Parallel()
		db := setupMongo(t)
		ctx := context.Background()
		repo := auctiondb.New(ctx, db)

		// Timestamp futuro: nunca atinge o critério de vencimento mesmo com AUCTION_INTERVAL=1s.
		id := uuid.NewString()
		require.NoError(t, repo.InsertAuctionForTest(ctx, id,
			"Future Item", "Cat", "a future auction for integration",
			auction.New, auction.Active, time.Now().Add(time.Hour).Unix()))

		monitorCtx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)
		repo.StartAuctionCloser(monitorCtx)

		require.Never(t, func() bool {
			found, err := repo.FindByID(ctx, id)
			return err != nil || found.Status != auction.Active
		}, 3*time.Second, 200*time.Millisecond,
			"auction %s status changed unexpectedly", id)
	})

	// Valida o ciclo de vida da goroutine do monitor: após o cancelamento do contexto,
	// a varredura para e um leilão vencido inserido em seguida permanece Active.
	t.Run("stops on context cancel", func(t *testing.T) {
		t.Parallel()
		db := setupMongo(t)
		ctx := context.Background()
		repo := auctiondb.New(ctx, db)

		// confirma que o monitor está vivo fechando um primeiro leilão vencido;
		// ao final do Eventually a goroutine está ociosa (bloqueada no select).
		aliveID := uuid.NewString()
		require.NoError(t, repo.InsertExpiredAuctionForTest(ctx, aliveID, time.Now().Add(-time.Hour).Unix()))

		monitorCtx, cancel := context.WithCancel(context.Background())
		repo.StartAuctionCloser(monitorCtx)
		waitForStatus(t, repo, aliveID, auction.Completed, 5*time.Second)

		// com o monitor ocioso, o cancelamento é observado imediatamente no select e a
		// goroutine encerra sem nova varredura.
		cancel()

		stoppedID := uuid.NewString()
		require.NoError(t, repo.InsertExpiredAuctionForTest(ctx, stoppedID, time.Now().Add(-time.Hour).Unix()))

		require.Never(t, func() bool {
			found, err := repo.FindByID(ctx, stoppedID)
			return err == nil && found.Status == auction.Completed
		}, 3*time.Second, 200*time.Millisecond,
			"auction %s should remain Active after the closer context was cancelled", stoppedID)
	})
}

// TestCreateAuction_ConcurrentClosers_Idempotent valida a idempotência quando
// scheduleAuctionClose e o monitor em background concorrem para fechar o mesmo
// leilão: apenas um executa a atualização; o segundo encontra status≠Active e é
// silencioso. O status final deve ser Completed sem oscilação.
func TestCreateAuction_ConcurrentClosers_Idempotent(t *testing.T) {
	t.Parallel()

	db := setupMongo(t)
	ctx := context.Background()
	repo := auctiondb.New(ctx, db)

	auctionEntity, err := auction.Create(
		"Concurrent Clock", "Decor", "A beautiful vintage wall clock", auction.New)
	require.Nil(t, err)
	// CreateAuction dispara scheduleAuctionClose (fecha após AUCTION_INTERVAL=1s).
	require.Nil(t, repo.Create(ctx, auctionEntity))

	// Monitor também tentará fechar o leilão após o ticket de 1s.
	monitorCtx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	repo.StartAuctionCloser(monitorCtx)

	// Aguarda o fechamento (qualquer fechador pode chegar primeiro).
	waitForStatus(t, repo, auctionEntity.ID, auction.Completed, 5*time.Second)

	// Após Completed, o status não deve oscilar: o segundo fechador deve ser no-op.
	require.Never(t, func() bool {
		found, err := repo.FindByID(ctx, auctionEntity.ID)
		return err != nil || found.Status != auction.Completed
	}, 3*time.Second, 200*time.Millisecond,
		"auction %s changed status after becoming Completed", auctionEntity.ID)
}
