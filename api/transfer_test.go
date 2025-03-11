package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	mockdb "github.com/techschool/simplebank/db/mock"
	db "github.com/techschool/simplebank/db/sqlc"
)

func TestTransferAPI(t *testing.T) {
	amount := int64(10)

	user1 := db.Account{
		ID:       1,
		Owner:    "owner1",
		Balance:  100,
		Currency: "USD",
	}

	user2 := db.Account{
		ID:       2,
		Owner:    "owner2",
		Balance:  100,
		Currency: "USD",
	}

	user3 := db.Account{
		ID:       3,
		Owner:    "owner3",
		Balance:  100,
		Currency: "EUR",
	}

	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"from_account_id": user1.ID,
				"to_account_id":   user2.ID,
				"amount":          amount,
				"currency":        "USD",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(user1.ID)).Times(1).Return(user1, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(user2.ID)).Times(1).Return(user2, nil)

				arg := db.TransferTxParams{
					FromAccountID: user1.ID,
					ToAccountID:   user2.ID,
					Amount:        amount,
				}

				result := db.TransferTxResult{
					Transfer: db.Transfer{
						ID:            1,
						FromAccountID: user1.ID,
						ToAccountID:   user2.ID,
						Amount:        amount,
					},
					FromAccount: user1,
					ToAccount:   user2,
					FromEntry: db.Entry{
						ID:        1,
						AccountID: user1.ID,
						Amount:    -amount,
					},
					ToEntry: db.Entry{
						ID:        2,
						AccountID: user2.ID,
						Amount:    amount,
					},
				}

				store.EXPECT().TransferTx(gomock.Any(), gomock.Eq(arg)).Times(1).Return(result, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchTransfer(t, recorder.Body)
			},
		},
		{
			name: "FromAccountNotFound",
			body: gin.H{
				"from_account_id": user1.ID,
				"to_account_id":   user2.ID,
				"amount":          amount,
				"currency":        "USD",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(user1.ID)).Times(1).Return(db.Account{}, sql.ErrNoRows)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(user2.ID)).Times(0)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "ToAccountNotFound",
			body: gin.H{
				"from_account_id": user1.ID,
				"to_account_id":   user2.ID,
				"amount":          amount,
				"currency":        "USD",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(user1.ID)).Times(1).Return(user1, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(user2.ID)).Times(1).Return(db.Account{}, sql.ErrNoRows)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "FromAccountCurrencyMismatch",
			body: gin.H{
				"from_account_id": user3.ID,
				"to_account_id":   user2.ID,
				"amount":          amount,
				"currency":        "USD",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(user3.ID)).Times(1).Return(user3, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(user2.ID)).Times(0)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "ToAccountCurrencyMismatch",
			body: gin.H{
				"from_account_id": user1.ID,
				"to_account_id":   user3.ID,
				"amount":          amount,
				"currency":        "USD",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(user1.ID)).Times(1).Return(user1, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(user3.ID)).Times(1).Return(user3, nil)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidCurrency",
			body: gin.H{
				"from_account_id": user1.ID,
				"to_account_id":   user2.ID,
				"amount":          amount,
				"currency":        "INVALID",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "NegativeAmount",
			body: gin.H{
				"from_account_id": user1.ID,
				"to_account_id":   user2.ID,
				"amount":          -amount,
				"currency":        "USD",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "GetAccountError",
			body: gin.H{
				"from_account_id": user1.ID,
				"to_account_id":   user2.ID,
				"amount":          amount,
				"currency":        "USD",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(1).Return(db.Account{}, sql.ErrConnDone)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "TransferTxError",
			body: gin.H{
				"from_account_id": user1.ID,
				"to_account_id":   user2.ID,
				"amount":          amount,
				"currency":        "USD",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(user1.ID)).Times(1).Return(user1, nil)
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(user2.ID)).Times(1).Return(user2, nil)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(1).Return(db.TransferTxResult{}, sql.ErrTxDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := NewServer(store)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/transfers"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func requireBodyMatchTransfer(t *testing.T, body *bytes.Buffer) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotTransfer db.TransferTxResult
	err = json.Unmarshal(data, &gotTransfer)
	require.NoError(t, err)

	require.NotEmpty(t, gotTransfer)
	require.NotZero(t, gotTransfer.Transfer.ID)
	require.NotZero(t, gotTransfer.FromEntry.ID)
	require.NotZero(t, gotTransfer.ToEntry.ID)
	require.NotZero(t, gotTransfer.FromAccount.ID)
	require.NotZero(t, gotTransfer.ToAccount.ID)
}
