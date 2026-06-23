package httperr_test

import (
	"net/http"
	"testing"

	"fullcycle-auction_go/configuration/httperr"
	"fullcycle-auction_go/internal/apperr"

	"github.com/stretchr/testify/require"
)

// TestConvertError valida o mapeamento de InternalError para o RestErr correto,
// incluindo o status HTTP — uma regressão aqui devolveria o código errado ao
// cliente.
func TestConvertError(t *testing.T) {
	cases := []struct {
		name     string
		in       *apperr.InternalError
		wantErr  string
		wantCode int
	}{
		{"bad request maps to 400", apperr.NewBadRequestError("bad"), "bad_request", http.StatusBadRequest},
		{"not found maps to 404", apperr.NewNotFoundError("missing"), "not_found", http.StatusNotFound},
		{"unknown maps to 500", apperr.NewInternalServerError("unexpected error"), "internal_server", http.StatusInternalServerError},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := httperr.ConvertError(tc.in)
			require.Equal(t, tc.wantCode, got.Code)
			require.Equal(t, tc.wantErr, got.Err)
			require.Equal(t, tc.in.Message, got.Message)
		})
	}
}
