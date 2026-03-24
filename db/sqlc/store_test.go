package db

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(testDB)

	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	oldBal1, err := decimal.NewFromString(account1.Balance)
	require.NoError(t, err)

	oldBal2, err := decimal.NewFromString(account2.Balance)
	require.NoError(t, err)

	n := int64(5)
	amount := decimal.New(10, 0)

	errs := make(chan error)
	results := make(chan TransferTxResult)

	for range n {
		go func() {
			result, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        amount.String(),
			})

			errs <- err
			results <- result
		}()
	}

	existed := make(map[int64]bool)

	for range n {
		err := <-errs
		require.NoError(t, err)

		result := <-results
		require.NotEmpty(t, result)

		transfer := result.Transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, account1.ID, transfer.FromAccountID)
		require.Equal(t, account2.ID, transfer.ToAccountID)
		require.Equal(t, amount.String(), transfer.Amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		_, err = store.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, account1.ID, fromEntry.AccountID)
		require.Equal(t, amount.Neg().String(), fromEntry.Amount)
		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, fromEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)

		toEntry := result.ToEntry
		require.NotEmpty(t, toEntry)
		require.Equal(t, account2.ID, toEntry.AccountID)
		require.Equal(t, amount.String(), toEntry.Amount)
		require.NotZero(t, toEntry.ID)
		require.NotZero(t, toEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), toEntry.ID)
		require.NoError(t, err)

		// TODO: check accounts' balance

		fromAccount := result.FromAccount
		require.NotEmpty(t, fromAccount)
		require.Equal(t, account1.ID, fromAccount.ID)

		toAccount := result.ToAccount
		require.NotEmpty(t, toAccount)
		require.Equal(t, account2.ID, toAccount.ID)

		newBal1, err := decimal.NewFromString(fromAccount.Balance)
		require.NoError(t, err)

		diff1 := oldBal1.Sub(newBal1)

		newBal2, err := decimal.NewFromString(toAccount.Balance)
		require.NoError(t, err)

		diff2 := newBal2.Sub(oldBal2)

		// INFO: oldBal is constant for every iteration, but newBal reflects the change, so the diff must be a power of amount

		require.True(t, diff1.Equal(diff2))
		require.True(t, diff1.IsPositive())
		require.True(t, diff1.Mod(amount).IsZero())

		k := diff1.Div(amount).IntPart()
		require.True(t, k >= 1 && k <= n)
		require.NotContains(t, existed, k)
		existed[k] = true
	}

	updatedAccount1, err := testQueries.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	updatedAccount2, err := testQueries.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)

	updatedAccount1Balance, err := decimal.NewFromString(updatedAccount1.Balance)
	require.NoError(t, err)

	updatedAccount2Balance, err := decimal.NewFromString(updatedAccount2.Balance)
	require.NoError(t, err)

	totalAmount := amount.Mul(decimal.New(5, 0))
	require.True(t, oldBal1.Sub(totalAmount).Equal(updatedAccount1Balance))
	require.True(t, oldBal2.Add(totalAmount).Equal(updatedAccount2Balance))
}
