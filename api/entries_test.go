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

func randomEntry() db.Entry {
	return db.Entry{
		ID:        util.RandomInt(1, 1000),
		AccountID: util.RandomInt(1, 1000),
		Amount:    util.RandomAmount(9999),
	}
}

func TestGetEntryAPI(t *testing.T) {
	entry := randomEntry()

	testCases := []struct {
		name          string
		entryID       int64
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:    "OK",
			entryID: entry.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEntry(gomock.Any(), gomock.Eq(entry.ID)).
					Times(1).
					Return(entry, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchEntry(t, recorder.Body, entry)
			},
		},
		{
			name:    "NotFound",
			entryID: entry.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEntry(gomock.Any(), gomock.Eq(entry.ID)).
					Times(1).
					Return(db.Entry{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:    "InternalError",
			entryID: entry.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEntry(gomock.Any(), gomock.Eq(entry.ID)).
					Times(1).
					Return(db.Entry{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:    "InvalidID",
			entryID: 0,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEntry(gomock.Any(), gomock.Any()).
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

			url := fmt.Sprintf("/entries/%d", tc.entryID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestListEntriesAPI(t *testing.T) {
	var entries []db.Entry
	for range 10 {
		entries = append(entries, randomEntry())
	}

	testCases := []struct {
		name          string
		query         url.Values
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK - All Entries",
			query: url.Values{
				"page": []string{"2"},
				"size": []string{"5"},
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListEntries(gomock.Any(), gomock.Eq(db.ListEntriesParams{
						Limit:  5,
						Offset: 5,
					})).
					Times(1).
					Return(entries[5:], nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchEntries(t, recorder.Body, entries[5:])
			},
		},
		{
			name: "OK - Account Entries",
			query: url.Values{
				"page":       []string{"2"},
				"size":       []string{"5"},
				"account_id": []string{fmt.Sprint(entries[0].AccountID)},
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccountEntries(gomock.Any(), gomock.Eq(db.ListAccountEntriesParams{
						AccountID: entries[0].AccountID,
						Limit:     5,
						Offset:    5,
					})).
					Times(1).
					Return(entries[5:], nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchEntries(t, recorder.Body, entries[5:])
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
					ListEntries(gomock.Any(), gomock.Any()).
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
					ListEntries(gomock.Any(), gomock.Eq(db.ListEntriesParams{
						Limit:  5,
						Offset: 5,
					})).
					Times(1).
					Return([]db.Entry{}, sql.ErrConnDone)
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

			url := url.URL{Path: "/entries", RawQuery: tc.query.Encode()}
			request, err := http.NewRequest(http.MethodGet, url.String(), nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func requireBodyMatchEntry(t *testing.T, body *bytes.Buffer, entry db.Entry) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotEntry db.Entry
	err = json.Unmarshal(data, &gotEntry)
	require.NoError(t, err)
	require.Equal(t, entry, gotEntry)
}

func requireBodyMatchEntries(t *testing.T, body *bytes.Buffer, entries []db.Entry) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotEntries []db.Entry
	err = json.Unmarshal(data, &gotEntries)
	require.NoError(t, err)
	require.Equal(t, entries, gotEntries)
}
