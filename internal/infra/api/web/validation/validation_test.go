package validation_test

import (
	"net/http"
	"testing"

	"fullcycle-auction_go/internal/infra/api/web/validation"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestValidateUUID(t *testing.T) {
	t.Parallel()

	t.Run("valid UUID returns nil", func(t *testing.T) {
		t.Parallel()
		err := validation.ValidateUUID(uuid.NewString(), "auctionId")
		require.Nil(t, err)
	})

	t.Run("invalid UUID returns bad request", func(t *testing.T) {
		t.Parallel()
		err := validation.ValidateUUID("not-a-uuid", "auctionId")
		require.NotNil(t, err)
		require.Equal(t, http.StatusBadRequest, err.Code)
		require.Len(t, err.Causes, 1)
		require.Equal(t, "auctionId", err.Causes[0].Field)
	})

	t.Run("empty string returns bad request", func(t *testing.T) {
		t.Parallel()
		err := validation.ValidateUUID("", "userId")
		require.NotNil(t, err)
		require.Equal(t, "userId", err.Causes[0].Field)
	})
}
