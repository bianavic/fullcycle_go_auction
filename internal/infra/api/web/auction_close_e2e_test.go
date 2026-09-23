//go:build integration

package web_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"fullcycle-auction_go/internal/app"
	"fullcycle-auction_go/internal/infra/api/web"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TestMain define um AUCTION_INTERVAL curto para que o teste de fechamento
// automático não precise esperar o padrão de produção (5 minutos).
func TestMain(m *testing.M) {
	_ = os.Setenv("AUCTION_INTERVAL", "2s")
	_ = os.Setenv("AUCTION_CLOSER_INTERVAL", "1s")
	os.Exit(m.Run())
}

// setupServer sobe um Mongo efêmero via Testcontainers, monta o mesmo wiring
// de produção (app.BuildDependencies + web.NewRouter) e expõe um
// httptest.Server real: as requisições passam por JSON binding, validação,
// controllers, use cases e repositório, exatamente como em produção.
func setupServer(t *testing.T) string {
	t.Helper()
	ctx := context.Background()

	container, err := mongodb.Run(ctx, "mongo:7")
	require.NoError(t, err, "failed to start mongodb container")
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			t.Logf("failed to terminate mongodb container: %v", err)
		}
	})

	uri, err := container.ConnectionString(ctx)
	require.NoError(t, err)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Disconnect(ctx) })
	require.NoError(t, client.Ping(ctx, nil))

	database := client.Database("auctions_e2e")

	appCtx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	userController, bidController, auctionController := app.BuildDependencies(appCtx, database)
	router := web.NewRouter(userController, bidController, auctionController)

	server := httptest.NewServer(router)
	t.Cleanup(server.Close)

	return server.URL
}

func createAuction(t *testing.T, baseURL string) {
	t.Helper()

	body, err := json.Marshal(map[string]any{
		"product_name": "Vintage Clock",
		"category":     "Decor",
		"description":  "a beautiful vintage wall clock from 1950",
		"condition":    1,
	})
	require.NoError(t, err)

	resp, err := http.Post(baseURL+"/auction", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusCreated, resp.StatusCode)
}

// findActiveAuctionID lista os leilões ativos via HTTP e retorna o id do
// único leilão criado pelo teste.
func findActiveAuctionID(t *testing.T, baseURL string) string {
	t.Helper()

	resp, err := http.Get(baseURL + "/auction?status=0")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var auctions []struct {
		ID     string `json:"id"`
		Status int    `json:"status"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&auctions))
	require.Len(t, auctions, 1, "expected exactly one active auction")

	return auctions[0].ID
}

// fetchAuctionStatus busca o leilão por id via HTTP e retorna seu status.
func fetchAuctionStatus(t *testing.T, baseURL, auctionID string) int {
	t.Helper()

	resp, err := http.Get(fmt.Sprintf("%s/auction/%s", baseURL, auctionID))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var auction struct {
		Status int `json:"status"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&auction))

	return auction.Status
}

// TestAuctionClosesAutomatically é o teste central pedido pelo desafio: um
// leilão criado via HTTP fecha sozinho após AUCTION_INTERVAL, sem qualquer
// chamada manual de fechamento — cobrindo a pilha completa (rota, binding,
// validação, use case, repositório e a goroutine de fechamento automático).
func TestAuctionClosesAutomatically(t *testing.T) {
	t.Parallel()

	baseURL := setupServer(t)

	createAuction(t, baseURL)
	auctionID := findActiveAuctionID(t, baseURL)

	require.Equal(t, 0, fetchAuctionStatus(t, baseURL, auctionID),
		"auction should start Active")

	require.Eventually(t, func() bool {
		return fetchAuctionStatus(t, baseURL, auctionID) == 1
	}, 6*time.Second, 200*time.Millisecond,
		"auction %s should close automatically after AUCTION_INTERVAL", auctionID)
}
