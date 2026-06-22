package bid

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestGetMaxBatchSize valida o parsing de MAX_BATCH_SIZE e o fallback para 5.
func TestGetMaxBatchSize(t *testing.T) {
	t.Run("default when unset", func(t *testing.T) {
		t.Setenv("MAX_BATCH_SIZE", "")
		require.Equal(t, 5, getMaxBatchSize())
	})
	t.Run("default when invalid", func(t *testing.T) {
		t.Setenv("MAX_BATCH_SIZE", "abc")
		require.Equal(t, 5, getMaxBatchSize())
	})
	t.Run("parses valid value", func(t *testing.T) {
		t.Setenv("MAX_BATCH_SIZE", "7")
		require.Equal(t, 7, getMaxBatchSize())
	})
}

// TestGetBatchInsertInterval valida o parsing de BATCH_INSERT_INTERVAL e o
// fallback para 3 minutos.
func TestGetBatchInsertInterval(t *testing.T) {
	t.Run("default when unset", func(t *testing.T) {
		t.Setenv("BATCH_INSERT_INTERVAL", "")
		require.Equal(t, 3*time.Minute, getBatchInsertInterval())
	})
	t.Run("default when invalid", func(t *testing.T) {
		t.Setenv("BATCH_INSERT_INTERVAL", "nonsense")
		require.Equal(t, 3*time.Minute, getBatchInsertInterval())
	})
	t.Run("parses valid value", func(t *testing.T) {
		t.Setenv("BATCH_INSERT_INTERVAL", "50ms")
		require.Equal(t, 50*time.Millisecond, getBatchInsertInterval())
	})
}
