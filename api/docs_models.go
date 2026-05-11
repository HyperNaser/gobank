package api

type ErrorResponse struct {
	Error string `json:"error"`
}

type AccountResponse struct {
	ID        int64  `json:"id"`
	Owner     string `json:"owner"`
	Balance   string `json:"balance"`
	Currency  string `json:"currency"`
	CreatedAt string `json:"created_at"`
	DeletedAt string `json:"deleted_at,omitempty"`
}

type EntryResponse struct {
	ID        int64  `json:"id"`
	AccountID int64  `json:"account_id"`
	Amount    string `json:"amount"`
	CreatedAt string `json:"created_at"`
}

type TransferResponse struct {
	ID            int64  `json:"id"`
	FromAccountID int64  `json:"from_account_id"`
	ToAccountID   int64  `json:"to_account_id"`
	Amount        string `json:"amount"`
	CreatedAt     string `json:"created_at"`
}

type TransferTxResultResponse struct {
	Transfer    TransferResponse `json:"transfer"`
	FromAccount AccountResponse  `json:"from_account"`
	ToAccount   AccountResponse  `json:"to_account"`
	FromEntry   EntryResponse    `json:"from_entry"`
	ToEntry     EntryResponse    `json:"to_entry"`
}
