// Package db is where all our db related operations are packaged
package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/shopspring/decimal"
)

type Store struct {
	*Queries
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db:      db,
		Queries: New(db),
	}
}

func (store *Store) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbError := tx.Rollback(); rbError != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbError)
		}
		return err
	}

	return tx.Commit()
}

type TransferTxParams struct {
	FromAccountID int64  `json:"from_account_id"`
	ToAccountID   int64  `json:"to_acocunt_id"`
	Amount        string `json:"amount"`
}

type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

func (store *Store) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams(arg))
		if err != nil {
			return err
		}

		amount, err := decimal.NewFromString(arg.Amount)
		if err != nil {
			return err
		}

		amountNegative := amount.Neg().String()

		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount:    amountNegative,
		})
		if err != nil {
			return err
		}

		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount:    arg.Amount,
		})
		if err != nil {
			return err
		}

		if arg.FromAccountID < arg.ToAccountID {
			result.FromAccount, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
				ID:     arg.FromAccountID,
				Amount: amountNegative,
			})
			if err != nil {
				return err
			}

			result.ToAccount, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
				ID:     arg.ToAccountID,
				Amount: arg.Amount,
			})
			if err != nil {
				return err
			}
		} else {
			result.ToAccount, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
				ID:     arg.ToAccountID,
				Amount: arg.Amount,
			})
			if err != nil {
				return err
			}

			result.FromAccount, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
				ID:     arg.FromAccountID,
				Amount: amountNegative,
			})
			if err != nil {
				return err
			}
		}

		return nil
	})

	return result, err
}

type DepositTxParams struct {
	AccountID int64  `json:"acocunt_id"`
	Amount    string `json:"amount"`
}

type DepositTxResult struct {
	Account Account `json:"account"`
	Entry   Entry   `json:"entry"`
}

func (store *Store) DepositTx(ctx context.Context, arg DepositTxParams) (DepositTxResult, error) {
	var result DepositTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		result.Entry, err = q.CreateEntry(ctx, CreateEntryParams(arg))
		if err != nil {
			return err
		}

		result.Account, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
			ID:     arg.AccountID,
			Amount: arg.Amount,
		})
		if err != nil {
			return err
		}

		return nil
	})

	return result, err
}
