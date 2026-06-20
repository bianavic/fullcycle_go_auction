package rest_err_test

import (
	"net/http"
	"testing"

	"fullcycle-auction_go/configuration/rest_err"
	"fullcycle-auction_go/internal/internal_error"

	"github.com/stretchr/testify/require"
)

// TestConvertError valida o mapeamento de InternalError para o RestErr correto,
// incluindo o status HTTP — uma regressão aqui devolveria o código errado ao
// cliente.
func TestConvertError(t *testing.T) {
	cases := []struct {
		name     string
		in       *internal_error.InternalError
		wantErr  string
		wantCode int
	}{
		{"bad request maps to 400", internal_error.NewBadRequestError("bad"), "bad_request", http.StatusBadRequest},
		{"not found maps to 404", internal_error.NewNotFoundError("missing"), "not_found", http.StatusNotFound},
		{"unknown maps to 500", internal_error.NewInternalServerError("unexpected error"), "internal_server", http.StatusInternalServerError},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := rest_err.ConvertError(tc.in)
			require.Equal(t, tc.wantCode, got.Code)
			require.Equal(t, tc.wantErr, got.Err)
			require.Equal(t, tc.in.Message, got.Message)
		})
	}
}
