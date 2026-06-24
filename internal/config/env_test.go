package config_test

import (
	"testing"
	"time"

	"fullcycle-auction_go/internal/config"

	"github.com/stretchr/testify/require"
)

func TestParseDuration(t *testing.T) {
	t.Run("valid env var", func(t *testing.T) {
		t.Setenv("TEST_DURATION", "30s")
		got := config.ParseDuration("TEST_DURATION", time.Minute)
		require.Equal(t, 30*time.Second, got)
	})

	t.Run("empty env var returns fallback", func(t *testing.T) {
		t.Setenv("TEST_DURATION_EMPTY", "")
		got := config.ParseDuration("TEST_DURATION_EMPTY", 5*time.Minute)
		require.Equal(t, 5*time.Minute, got)
	})

	t.Run("missing env var returns fallback", func(t *testing.T) {
		got := config.ParseDuration("TEST_DURATION_MISSING_KEY", 10*time.Second)
		require.Equal(t, 10*time.Second, got)
	})

	t.Run("invalid format returns fallback", func(t *testing.T) {
		t.Setenv("TEST_DURATION_BAD", "not-a-duration")
		got := config.ParseDuration("TEST_DURATION_BAD", 2*time.Minute)
		require.Equal(t, 2*time.Minute, got)
	})
}
