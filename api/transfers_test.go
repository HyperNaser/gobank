package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	mockdb "github.com/HyperNaser/gobank/db/mock"
	db "github.com/HyperNaser/gobank/db/sqlc"
	"github.com/HyperNaser/gobank/util"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func randomTransfer() db.Transfer {
	return db.Transfer{
		ID:            util.RandomInt(1, 1000),
		FromAccountID: util.RandomInt(1, 1000),
		ToAccountID:   util.RandomInt(1001, 2001),
		Amount:        util.RandomAmount(9999),
	}
}

func TestGetTransferAPI(t *testing.T) {
	transfer := randomTransfer()

	testCases := []struct {
		name          string
		transferID    int64
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:       "OK",
			transferID: transfer.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetTransfer(gomock.Any(), gomock.Eq(transfer.ID)).
					Times(1).
					Return(transfer, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchTransfer(t, recorder.Body, transfer)
			},
		},
		{
			name:       "NotFound",
			transferID: transfer.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetTransfer(gomock.Any(), gomock.Eq(transfer.ID)).
					Times(1).
					Return(db.Transfer{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:       "InternalError",
			transferID: transfer.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetTransfer(gomock.Any(), gomock.Eq(transfer.ID)).
					Times(1).
					Return(db.Transfer{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:       "InvalidID",
			transferID: 0,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetTransfer(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := NewServer(store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/transfers/%d", tc.transferID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestListTransfersAPI(t *testing.T) {
	var transfers []db.Transfer
	for range 10 {
		transfers = append(transfers, randomTransfer())
	}

	testCases := []struct {
		name          string
		query         url.Values
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK - All Transfers",
			query: url.Values{
				"page": []string{"2"},
				"size": []string{"5"},
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListTransfers(gomock.Any(), gomock.Eq(db.ListTransfersParams{
						Limit:  5,
						Offset: 5,
					})).
					Times(1).
					Return(transfers[5:], nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchTransfers(t, recorder.Body, transfers[5:])
			},
		},
		{
			name: "OK - Account Transfers",
			query: url.Values{
				"page":       []string{"2"},
				"size":       []string{"5"},
				"account_id": []string{fmt.Sprint(transfers[0].FromAccountID)},
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccountTransfers(gomock.Any(), gomock.Eq(db.ListAccountTransfersParams{
						FromAccountID: transfers[0].FromAccountID,
						Limit:         5,
						Offset:        5,
					})).
					Times(1).
					Return(transfers[5:], nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchTransfers(t, recorder.Body, transfers[5:])
			},
		},
		{
			name: "InvalidQuery",
			query: url.Values{
				"page": []string{"-2"},
				"size": []string{"5"},
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListTransfers(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InternalError",
			query: url.Values{
				"page": []string{"2"},
				"size": []string{"5"},
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListTransfers(gomock.Any(), gomock.Eq(db.ListTransfersParams{
						Limit:  5,
						Offset: 5,
					})).
					Times(1).
					Return([]db.Transfer{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := NewServer(store)
			recorder := httptest.NewRecorder()

			url := url.URL{Path: "/transfers", RawQuery: tc.query.Encode()}
			request, err := http.NewRequest(http.MethodGet, url.String(), nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func requireBodyMatchTransfer(t *testing.T, body *bytes.Buffer, transfer db.Transfer) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotTransfer db.Transfer
	err = json.Unmarshal(data, &gotTransfer)
	require.NoError(t, err)
	require.Equal(t, transfer, gotTransfer)
}

func requireBodyMatchTransfers(t *testing.T, body *bytes.Buffer, transfers []db.Transfer) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotTransfers []db.Transfer
	err = json.Unmarshal(data, &gotTransfers)
	require.NoError(t, err)
	require.Equal(t, transfers, gotTransfers)
}
