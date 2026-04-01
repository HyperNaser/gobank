package db

import (
	"context"
	"testing"
	"time"

	"github.com/HyperNaser/gobank/util"
	"github.com/stretchr/testify/require"
)

func createRandomUser(t *testing.T) User {
	arg := CreateUserParams{
		Username:       util.RandomOwner(),
		HashedPassword: "password",
		FullName:       util.RandomOwner(),
		Email:          util.RandomEmail(),
	}

	user, err := testQueries.CreateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, arg.Username, user.Username)
	require.Equal(t, arg.FullName, user.FullName)
	require.Equal(t, arg.HashedPassword, user.HashedPassword)
	require.Equal(t, arg.Email, user.Email)

	require.NotZero(t, user.CreatedAt)
	require.Zero(t, user.PasswordChangedAt)

	return user
}

func TestCreateUser(t *testing.T) {
	createRandomUser(t)
}

func TestGetUser(t *testing.T) {
	user1 := createRandomUser(t)
	user2, err := testQueries.GetUser(context.Background(), user1.Username)
	require.NoError(t, err)
	require.NotEmpty(t, user2)

	require.Equal(t, user1.Username, user2.Username)
	require.Equal(t, user1.FullName, user2.FullName)
	require.Equal(t, user1.HashedPassword, user2.HashedPassword)
	require.Equal(t, user1.Email, user2.Email)
	require.WithinDuration(t, user1.CreatedAt, user2.CreatedAt, time.Second)
	require.WithinDuration(t, user1.PasswordChangedAt, user2.PasswordChangedAt, time.Second)
}

// func TestUpdateUser(t *testing.T) {
// 	user1 := createRandomAccount(t)
//
// 	arg := UpdateUserParams{
// 		ID:      user1.ID,
// 		Balance: util.RandomMoney(-5000, 5000),
// 	}
//
// 	account2, err := testQueries.UpdateAccount(context.Background(), arg)
// 	require.NoError(t, err)
// 	require.NotEmpty(t, account2)
//
// 	require.Equal(t, user1.ID, account2.ID)
// 	require.Equal(t, user1.Owner, account2.Owner)
// 	require.Equal(t, arg.Balance, account2.Balance)
// 	require.Equal(t, user1.Currency, account2.Currency)
// 	require.WithinDuration(t, user1.CreatedAt, account2.CreatedAt, time.Second)
// }
//
// func TestDeleteAccount(t *testing.T) {
// 	account1 := createRandomAccount(t)
// 	err := testQueries.DeleteAccount(context.Background(), account1.ID)
// 	require.NoError(t, err)
//
// 	account2, err := testQueries.GetAccount(context.Background(), account1.ID)
// 	require.Error(t, err)
// 	require.EqualError(t, err, sql.ErrNoRows.Error())
// 	require.Empty(t, account2)
// }
//
// func TestListAccounts(t *testing.T) {
// 	for range 10 {
// 		createRandomAccount(t)
// 	}
//
// 	arg := ListAccountsParams{
// 		Limit:  5,
// 		Offset: 5,
// 	}
//
// 	accounts, err := testQueries.ListAccounts(context.Background(), arg)
// 	require.NoError(t, err)
// 	require.Len(t, accounts, 5)
//
// 	for _, account := range accounts {
// 		require.NotEmpty(t, account)
// 	}
// }
