package web_test

import (
	"math/big"
	"net/http"
	"testing"

	"chainlink/core/internal/cltest"
	"chainlink/core/store/models"
	"chainlink/core/web"

	"github.com/manyminds/api2go/jsonapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTxAttemptsController_Index_Success(t *testing.T) {
	t.Parallel()

	app, cleanup := cltest.NewApplicationWithKey(t)
	defer cleanup()

	ethMock := app.MockCallerSubscriberClient()
	ethMock.Context("app.Start()", func(ethMock *cltest.EthMock) {
		ethMock.Register("eth_getTransactionCount", "0x100")
	})

	require.NoError(t, app.Start())
	store := app.GetStore()
	client := app.NewHTTPClient()

	from := cltest.GetAccountAddress(t, store)
	tx := cltest.CreateTx(t, store, from, 1)
	_, err := store.AddTxAttempt(tx, tx.EthTx(big.NewInt(2)), 2)
	require.NoError(t, err)
	_, err = store.AddTxAttempt(tx, tx.EthTx(big.NewInt(3)), 3)
	require.NoError(t, err)

	resp, cleanup := client.Get("/v2/tx_attempts?size=2")
	defer cleanup()
	cltest.AssertServerResponse(t, resp, http.StatusOK)

	var links jsonapi.Links
	var attempts []models.TxAttempt
	body := cltest.ParseResponseBody(t, resp)
	require.NoError(t, web.ParsePaginatedResponse(body, &attempts, &links))
	assert.NotEmpty(t, links["next"].Href)
	assert.Empty(t, links["prev"].Href)

	assert.Len(t, attempts, 2)
	assert.Equal(t, uint64(3), attempts[0].SentAt, "expected tx attempts order by sentAt descending")
	assert.Equal(t, uint64(2), attempts[1].SentAt, "expected tx attempts order by sentAt descending")
}

func TestTxAttemptsController_Index_Error(t *testing.T) {
	t.Parallel()

	app, cleanup := cltest.NewApplicationWithKey(t)
	defer cleanup()
	ethMock := app.MockCallerSubscriberClient()
	ethMock.Register("eth_getTransactionCount", "0x100")
	require.NoError(t, app.Start())

	client := app.NewHTTPClient()
	resp, cleanup := client.Get("/v2/tx_attempts?size=TrainingDay")
	defer cleanup()
	cltest.AssertServerResponse(t, resp, 422)
}
