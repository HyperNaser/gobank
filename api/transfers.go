package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	db "github.com/HyperNaser/gobank/db/sqlc"
	"github.com/HyperNaser/gobank/token"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type getTransferRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

// getTransfer retrieves a transfer by ID when the authenticated user participates in it.
// @Summary Get transfer
// @Description Get a transfer by ID when the authenticated user is sender or receiver.
// @Tags transfers
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Transfer ID"
// @Success 200 {object} TransferResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /transfers/{id} [get]
func (server *Server) getTransfer(ctx *gin.Context) {
	var req getTransferRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	transfer, err := server.store.GetTransfer(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	fromAccount, err := server.store.GetAccount(ctx, transfer.FromAccountID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if authPayload.Username != fromAccount.Owner {
		toAccount, err := server.store.GetAccount(ctx, transfer.ToAccountID)
		if err != nil {
			if err == sql.ErrNoRows {
				ctx.JSON(http.StatusNotFound, errorResponse(err))
				return
			}
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}

		if authPayload.Username != toAccount.Owner {
			err := errors.New("transfer doesn't belong to authenticated user")
			ctx.JSON(http.StatusUnauthorized, errorResponse(err))
			return
		}
	}

	ctx.JSON(http.StatusOK, transfer)
}

type listAccountTransfersRequest struct {
	AccountID int64 `form:"account_id" binding:"required,min=1"`
	Page      int32 `form:"page" binding:"required,min=1"`
	Size      int32 `form:"size" binding:"required,min=5,max=10"`
}

// listTransfers returns paginated transfers for an account owned by the authenticated user.
// @Summary List transfers
// @Description List transfers for a specific account with pagination.
// @Tags transfers
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param account_id query int true "Account ID"
// @Param page query int true "Page number"
// @Param size query int true "Page size"
// @Success 200 {array} TransferResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /transfers [get]
func (server *Server) listTransfers(ctx *gin.Context) {
	var req listAccountTransfersRequest

	if err := ctx.ShouldBindWith(&req, binding.Query); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var (
		transfers []db.Transfer
		err       error
	)

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	account, err := server.store.GetAccount(ctx, req.AccountID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if authPayload.Username != account.Owner {
		err := errors.New("transfers don't belong to authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	arg := db.ListAccountTransfersParams{
		FromAccountID: req.AccountID,
		Limit:         req.Size,
		Offset:        (req.Page - 1) * req.Size,
	}
	transfers, err = server.store.ListAccountTransfers(ctx, arg)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, transfers)
}

type transferRequest struct {
	FromAccountID int64  `json:"from_account_id" binding:"required,min=1"`
	ToAccountID   int64  `json:"to_account_id" binding:"required,min=1"`
	Amount        string `json:"amount" binding:"required,gt=0"`
	Currency      string `json:"currency" binding:"required,currency"`
}

// createTransfer executes a transfer between two accounts owned by the authenticated user.
// @Summary Create transfer
// @Description Create a transfer from one account to another in a supported currency.
// @Tags transfers
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param transfer body transferRequest true "Transfer request"
// @Success 200 {object} TransferTxResultResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /transfers [post]
func (server *Server) createTransfer(ctx *gin.Context) {
	var req transferRequest
	if err := ctx.ShouldBindWith(&req, binding.JSON); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	fromAccount, valid := server.validAccount(ctx, req.FromAccountID, req.Currency)
	if !valid {
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if fromAccount.Owner != authPayload.Username {
		err := errors.New("from account doesn't belong to authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	_, valid = server.validAccount(ctx, req.ToAccountID, req.Currency)
	if !valid {
		return
	}

	arg := db.TransferTxParams{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
	}

	result, err := server.store.TransferTx(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (server *Server) validAccount(ctx *gin.Context, accountID int64, currency string) (db.Account, bool) {
	account, err := server.store.GetAccount(ctx, accountID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return account, false
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return account, false
	}

	if account.Currency != currency {
		err := fmt.Errorf("account [%d] currency mismatch: %s vs %s", accountID, account.Currency, currency)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return account, false
	}

	return account, true
}
