package bid_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"fullcycle-auction_go/internal/apperr"
	"fullcycle-auction_go/internal/entity/bid"
	biduc "fullcycle-auction_go/internal/usecase/bid"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// fakeBidRepo é um stub de BidEntityRepository que registra os lotes
// recebidos por CreateBid e devolve respostas configuráveis para as buscas.
type fakeBidRepo struct {
	mu             sync.Mutex
	createdBatches [][]bid.Bid
	createErr      *apperr.InternalError

	findBids   []bid.Bid
	findErr    *apperr.InternalError
	winning    *bid.Bid
	winningErr *apperr.InternalError
}

func (f *fakeBidRepo) Create(ctx context.Context, bids []bid.Bid) *apperr.InternalError {
	f.mu.Lock()
	defer f.mu.Unlock()
	batch := make([]bid.Bid, len(bids))
	copy(batch, bids)
	f.createdBatches = append(f.createdBatches, batch)
	return f.createErr
}

func (f *fakeBidRepo) FindByAuctionID(ctx context.Context, auctionID string) ([]bid.Bid, *apperr.InternalError) {
	return f.findBids, f.findErr
}

func (f *fakeBidRepo) FindWinningByAuctionID(ctx context.Context, auctionID string) (*bid.Bid, *apperr.InternalError) {
	return f.winning, f.winningErr
}

func (f *fakeBidRepo) batches() [][]bid.Bid {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([][]bid.Bid, len(f.createdBatches))
	copy(out, f.createdBatches)
	return out
}

func (f *fakeBidRepo) totalBids() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	n := 0
	for _, b := range f.createdBatches {
		n += len(b)
	}
	return n
}

func TestCreateBid_FlushBehavior(t *testing.T) {
	t.Run("flushes on max batch size", func(t *testing.T) {
		t.Setenv("MAX_BATCH_SIZE", "2")
		t.Setenv("BATCH_INSERT_INTERVAL", "1m")

		repo := &fakeBidRepo{}
		uc := biduc.New(repo)
		ctx := context.Background()

		auctionID := uuid.NewString()
		require.Nil(t, uc.CreateBid(ctx, biduc.InputDTO{
			UserID: uuid.NewString(), AuctionID: auctionID, Amount: 100}))
		require.Nil(t, uc.CreateBid(ctx, biduc.InputDTO{
			UserID: uuid.NewString(), AuctionID: auctionID, Amount: 200}))

		require.Eventually(t, func() bool {
			return repo.totalBids() == 2
		}, 2*time.Second, 20*time.Millisecond, "expected batch flush of 2 bids")

		batches := repo.batches()
		require.Len(t, batches, 1)
		require.Len(t, batches[0], 2)
	})

	t.Run("flushes on timer expiry", func(t *testing.T) {
		t.Setenv("MAX_BATCH_SIZE", "10")
		t.Setenv("BATCH_INSERT_INTERVAL", "100ms")

		repo := &fakeBidRepo{}
		uc := biduc.New(repo)
		ctx := context.Background()

		require.Nil(t, uc.CreateBid(ctx, biduc.InputDTO{
			UserID: uuid.NewString(), AuctionID: uuid.NewString(), Amount: 100}))

		require.Eventually(t, func() bool {
			return repo.totalBids() == 1
		}, 2*time.Second, 20*time.Millisecond, "expected timer-based flush of 1 bid")
	})

	// Valida o fix: o flush por timer só ocorre quando há bids acumulados (if len(batch) > 0
	// em create.go). A asserção é sobre o número de chamadas (batches()), não totalBids():
	// um lote vazio contribui 0 bids de qualquer forma, então só a contagem de chamadas
	// distingue o código com guard do código sem guard.
	t.Run("empty batch does not flush on timer", func(t *testing.T) {
		t.Setenv("MAX_BATCH_SIZE", "10")
		t.Setenv("BATCH_INSERT_INTERVAL", "50ms")

		repo := &fakeBidRepo{}
		_ = biduc.New(repo) // inicia a goroutine; nenhum bid enfileirado

		require.Never(t, func() bool {
			return len(repo.batches()) > 0
		}, 300*time.Millisecond, 20*time.Millisecond,
			"timer must not flush an empty batch")
	})
}

func TestCreateBid_Validation(t *testing.T) {
	t.Parallel()

	// barrado antes de enfileirar.
	t.Run("invalid user ID returns bad request", func(t *testing.T) {
		t.Parallel()
		repo := &fakeBidRepo{}
		uc := biduc.New(repo)

		err := uc.CreateBid(context.Background(), biduc.InputDTO{
			UserID: "not-a-uuid", AuctionID: uuid.NewString(), Amount: 100})
		require.NotNil(t, err)
		require.Equal(t, "bad_request", err.Err)
		require.Equal(t, 0, repo.totalBids())
	})

	t.Run("negative amount returns bad request", func(t *testing.T) {
		t.Parallel()
		repo := &fakeBidRepo{}
		uc := biduc.New(repo)

		err := uc.CreateBid(context.Background(), biduc.InputDTO{
			UserID: uuid.NewString(), AuctionID: uuid.NewString(), Amount: -5})
		require.NotNil(t, err)
		require.Equal(t, "bad_request", err.Err)
		require.Equal(t, 0, repo.totalBids())
	})
}
