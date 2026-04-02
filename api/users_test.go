package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	mockdb "github.com/HyperNaser/gobank/db/mock"
	db "github.com/HyperNaser/gobank/db/sqlc"
	"github.com/HyperNaser/gobank/util"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/lib/pq/pqerror"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type eqCreateUserParams struct {
	arg      db.CreateUserParams
	password string
}

func EqCreateUserParams(arg db.CreateUserParams, password string) gomock.Matcher {
	return eqCreateUserParams{arg, password}
}

func (e eqCreateUserParams) Matches(x any) bool {
	if arg, ok := x.(db.CreateUserParams); ok {
		err := util.CheckPassword(e.password, arg.HashedPassword)
		if err != nil {
			return false
		}
		e.arg.HashedPassword = arg.HashedPassword
		return reflect.DeepEqual(e.arg, arg)
	}

	return false
}

func (e eqCreateUserParams) String() string {
	return fmt.Sprintf("matches arg %v and password %s", e.arg, e.password)
}

func randomUserAndPassword() (db.User, string) {
	return db.User{
		Username: util.RandomOwner(),
		FullName: util.RandomOwner(),
		Email:    util.RandomEmail(),
	}, util.RandomString(12)
}

func TestCreateUserAPI(t *testing.T) {
	user, password := randomUserAndPassword()

	arg := db.CreateUserParams{
		Username:       user.Username,
		FullName:       user.FullName,
		HashedPassword: user.HashedPassword,
		Email:          user.Email,
	}

	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"username":  arg.Username,
				"full_name": arg.FullName,
				"password":  password,
				"email":     arg.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), EqCreateUserParams(arg, password)).
					Times(1).
					Return(user, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchUser(t, recorder.Body, user)
			},
		},
		{
			name: "InvalidUsername",
			body: gin.H{
				"username":  "Inval1d Username)_!",
				"full_name": arg.FullName,
				"password":  password,
				"email":     arg.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "PasswordTooShort",
			body: gin.H{
				"username":  arg.Username,
				"full_name": arg.FullName,
				"password":  "x1231",
				"email":     arg.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidEmail",
			body: gin.H{
				"username":  arg.Username,
				"full_name": arg.FullName,
				"password":  password,
				"email":     "invalidEmail",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "DuplicateUsername",
			body: gin.H{
				"username":  arg.Username,
				"full_name": arg.FullName,
				"password":  password,
				"email":     arg.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, &pq.Error{Code: pqerror.UniqueViolation})
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"username":  arg.Username,
				"full_name": arg.FullName,
				"password":  password,
				"email":     arg.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, sql.ErrConnDone)
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

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/users"

			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func requireBodyMatchUser(t *testing.T, body *bytes.Buffer, user db.User) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var response createUserResponse
	err = json.Unmarshal(data, &response)
	require.NoError(t, err)

	require.Equal(t, createUserResponse{
		Username: user.Username,
		FullName: user.FullName,
		Email:    user.Email,
	}, response)
}
