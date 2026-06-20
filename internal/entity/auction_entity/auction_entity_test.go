package auction_entity_test

import (
	"testing"

	"fullcycle-auction_go/internal/entity/auction_entity"

	"github.com/stretchr/testify/require"
)

// TestAuction_Validate cobre os ramos inequívocos de Validate (productName,
// category e caminho válido). Casos que dependem da precedência booleana entre
// Description e Condition foram omitidos de propósito: o comportamento atual é um
// bug marcado com TODO em auction_entity.go e não deve ser travado por testes.
func TestAuction_Validate(t *testing.T) {
	cases := []struct {
		name        string
		productName string
		category    string
		description string
		condition   auction_entity.ProductCondition
		wantErr     bool
	}{
		{"all valid", "Clock", "Decor", "A long enough description", auction_entity.New, false},
		{"product name too short", "C", "Decor", "A long enough description", auction_entity.New, true},
		{"category too short", "Clock", "De", "A long enough description", auction_entity.New, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := auction_entity.Auction{
				ProductName: tc.productName,
				Category:    tc.category,
				Description: tc.description,
				Condition:   tc.condition,
				Status:      auction_entity.Active,
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

// TestCreateAuction_Valid confirma que o construtor gera Id, marca Active e valida.
func TestCreateAuction_Valid(t *testing.T) {
	t.Parallel()

	a, err := auction_entity.CreateAuction(
		"Clock", "Decor", "A long enough description", auction_entity.New)
	require.Nil(t, err)
	require.NotEmpty(t, a.Id)
	require.Equal(t, auction_entity.Active, a.Status)
}

// TestCreateAuction_Invalid confirma que entradas inválidas não produzem entidade.
func TestCreateAuction_Invalid(t *testing.T) {
	t.Parallel()

	a, err := auction_entity.CreateAuction(
		"C", "Decor", "A long enough description", auction_entity.New)
	require.NotNil(t, err)
	require.Nil(t, a)
}
