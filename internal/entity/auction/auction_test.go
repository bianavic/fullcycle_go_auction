package auction_test

import (
	"testing"

	"fullcycle-auction_go/internal/entity/auction"

	"github.com/stretchr/testify/require"
)

func TestAuction_Validate(t *testing.T) {
	cases := []struct {
		name        string
		productName string
		category    string
		description string
		condition   auction.ProductCondition
		wantErr     bool
	}{
		{"all valid", "Clock", "Decor", "A long enough description", auction.New, false},
		{"product name too short", "C", "Decor", "A long enough description", auction.New, true},
		{"category too short", "Clock", "De", "A long enough description", auction.New, true},
		{"description too short", "Clock", "Decor", "Short", auction.New, true},
		{"invalid condition", "Clock", "Decor", "A long enough description", 0, true},
		{"valid with used condition", "Clock", "Decor", "A long enough description", auction.Used, false},
		{"valid with refurbished condition", "Clock", "Decor", "A long enough description", auction.Refurbished, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := auction.Auction{
				ProductName: tc.productName,
				Category:    tc.category,
				Description: tc.description,
				Condition:   tc.condition,
				Status:      auction.Active,
			}

			err := a.Validate()
			if tc.wantErr {
				require.NotNil(t, err)
				require.Equal(t, "bad_request", err.Err)
				return
			}
			require.Nil(t, err)
		})
	}
}

func TestCreateAuction(t *testing.T) {
	t.Parallel()

	t.Run("valid input", func(t *testing.T) {
		t.Parallel()
		a, err := auction.Create(
			"Clock", "Decor", "A long enough description", auction.New)
		require.Nil(t, err)
		require.NotEmpty(t, a.ID)
		require.Equal(t, auction.Active, a.Status)
	})

	t.Run("invalid input", func(t *testing.T) {
		t.Parallel()
		a, err := auction.Create(
			"C", "Decor", "A long enough description", auction.New)
		require.NotNil(t, err)
		require.Nil(t, a)
	})
}
