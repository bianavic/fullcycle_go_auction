package bid_entity_test

import (
	"testing"

	"fullcycle-auction_go/internal/entity/bid_entity"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// TestCreateBid_Validation cobre a validação de UserId/AuctionId (UUID) e Amount.
func TestCreateBid_Validation(t *testing.T) {
	cases := []struct {
		name      string
		userId    string
		auctionId string
		amount    float64
		wantErr   bool
	}{
		{"valid", uuid.NewString(), uuid.NewString(), 100, false},
		{"invalid user id", "not-a-uuid", uuid.NewString(), 100, true},
		{"invalid auction id", uuid.NewString(), "not-a-uuid", 100, true},
		{"zero amount", uuid.NewString(), uuid.NewString(), 0, true},
		{"negative amount", uuid.NewString(), uuid.NewString(), -5, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			b, err := bid_entity.CreateBid(tc.userId, tc.auctionId, tc.amount)
			if tc.wantErr {
				require.NotNil(t, err)
				require.Nil(t, b)
				require.Equal(t, "bad_request", err.Err)
				return
			}
			require.Nil(t, err)
			require.NotEmpty(t, b.Id)
			require.Equal(t, tc.amount, b.Amount)
		})
	}
}
