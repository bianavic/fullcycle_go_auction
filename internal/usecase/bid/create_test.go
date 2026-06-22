package bid_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"fullcycle-auction_go/internal/entity/bid"
	"fullcycle-auction_go/internal/internal_error"
	biduc "fullcycle-auction_go/internal/usecase/bid"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// fakeBidRepo é um stub de BidEntityRepository que registra os lotes
// recebidos por CreateBid e devolve respostas configuráveis para as buscas.
type fakeBidRepo struct {
	mu             sync.Mutex
	createdBatches [][]bid.Bid
	createErr      *internal_error.InternalError

	findBids   []bid.Bid
	findErr    *internal_error.InternalError
	winning    *bid.Bid
	winningErr *internal_error.InternalError
}

func (f *fakeBidRepo) CreateBid(ctx context.Context, bids []bid.Bid) *internal_error.InternalError {
	f.mu.Lock()
	defer f.mu.Unlock()
	batch := make([]bid.Bid, len(bids))
	copy(batch, bids)
	f.createdBatches = append(f.createdBatches, batch)
	return f.createErr
}

func (f *fakeBidRepo) FindBidByAuctionID(ctx context.Context, auctionID string) ([]bid.Bid, *internal_error.InternalError) {
	return f.findBids, f.findErr
}

func (f *fakeBidRepo) FindWinningBidByAuctionID(ctx context.Context, auctionID string) (*bid.Bid, *internal_error.InternalError) {
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

// TestCreateBid_FlushesOnMaxBatchSize valida que o lote é persistido assim que
// atinge maxBatchSize (também cobre o caminho de input válido enfileirado).
func TestCreateBid_FlushesOnMaxBatchSize(t *testing.T) {
	t.Setenv("MAX_BATCH_SIZE", "2")
	t.Setenv("BATCH_INSERT_INTERVAL", "1m")

	repo := &fakeBidRepo{}
	uc := biduc.New(repo)
	ctx := context.Background()

	auctionID := uuid.NewString()
	require.Nil(t, uc.CreateBid(ctx, biduc.BidInputDTO{
		UserID: uuid.NewString(), AuctionID: auctionID, Amount: 100}))
	require.Nil(t, uc.CreateBid(ctx, biduc.BidInputDTO{
		UserID: uuid.NewString(), AuctionID: auctionID, Amount: 200}))

	require.Eventually(t, func() bool {
		return repo.totalBids() == 2
	}, 2*time.Second, 20*time.Millisecond, "expected batch flush of 2 bids")

	batches := repo.batches()
	require.Len(t, batches, 1)
	require.Len(t, batches[0], 2)
}

// TestCreateBid_FlushesOnTimerExpiry valida que um lote abaixo de maxBatchSize é
// persistido quando o timer expira.
func TestCreateBid_FlushesOnTimerExpiry(t *testing.T) {
	t.Setenv("MAX_BATCH_SIZE", "10")
	t.Setenv("BATCH_INSERT_INTERVAL", "100ms")

	repo := &fakeBidRepo{}
	uc := biduc.New(repo)
	ctx := context.Background()

	require.Nil(t, uc.CreateBid(ctx, biduc.BidInputDTO{
		UserID: uuid.NewString(), AuctionID: uuid.NewString(), Amount: 100}))

	require.Eventually(t, func() bool {
		return repo.totalBids() == 1
	}, 2*time.Second, 20*time.Millisecond, "expected timer-based flush of 1 bid")
}

// TestCreateBid_EmptyBatch_TimerDoesNotFlush guarda o fix do commit 8: o flush por
// timer só ocorre quando há bids acumulados (if len(batch) > 0 em
// create.go). Sem nenhum CreateBid, o timer expira repetidamente e
// nenhum lote (nem vazio) deve ser enviado ao repositório. A asserção é sobre o
// NÚMERO DE CHAMADAS (batches()), não totalBids(): um lote vazio contribui 0 bids
// de qualquer forma, então só a contagem de chamadas distingue o código com guard
// do código sem guard.
func TestCreateBid_EmptyBatch_TimerDoesNotFlush(t *testing.T) {
	t.Setenv("MAX_BATCH_SIZE", "10")
	t.Setenv("BATCH_INSERT_INTERVAL", "50ms")

	repo := &fakeBidRepo{}
	_ = biduc.New(repo) // inicia a goroutine; nenhum bid enfileirado

	require.Never(t, func() bool {
		return len(repo.batches()) > 0
	}, 300*time.Millisecond, 20*time.Millisecond,
		"timer must not flush an empty batch")
}

// TestCreateBid_InvalidUserId_ReturnsBadRequest valida que um userID inválido é
// barrado antes de enfileirar.
func TestCreateBid_InvalidUserId_ReturnsBadRequest(t *testing.T) {
	repo := &fakeBidRepo{}
	uc := biduc.New(repo)

	err := uc.CreateBid(context.Background(), biduc.BidInputDTO{
		UserID: "not-a-uuid", AuctionID: uuid.NewString(), Amount: 100})
	require.NotNil(t, err)
	require.Equal(t, "bad_request", err.Err)
	require.Equal(t, 0, repo.totalBids())
}

// TestCreateBid_NegativeAmount_ReturnsBadRequest valida o guard de amount <= 0.
func TestCreateBid_NegativeAmount_ReturnsBadRequest(t *testing.T) {
	repo := &fakeBidRepo{}
	uc := biduc.New(repo)

	err := uc.CreateBid(context.Background(), biduc.BidInputDTO{
		UserID: uuid.NewString(), AuctionID: uuid.NewString(), Amount: -5})
	require.NotNil(t, err)
	require.Equal(t, "bad_request", err.Err)
	require.Equal(t, 0, repo.totalBids())
}
