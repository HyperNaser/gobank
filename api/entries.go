package api

import (
	"database/sql"
	"errors"
	"net/http"

	db "github.com/HyperNaser/gobank/db/sqlc"
	"github.com/HyperNaser/gobank/token"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type getEntryRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

// getEntry retrieves a ledger entry by ID for the authenticated user.
// @Summary Get entry
// @Description Get a single entry by ID when it belongs to the authenticated user.
// @Tags entries
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Entry ID"
// @Success 200 {object} EntryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /entries/{id} [get]
func (server *Server) getEntry(ctx *gin.Context) {
	var req getEntryRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	entry, err := server.store.GetEntry(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	account, err := server.store.GetAccount(ctx, entry.AccountID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if account.Owner != authPayload.Username {
		err := errors.New("entry doesn't belong to authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, entry)
}

type listEntriesRequest struct {
	AccountID int64 `form:"account_id" binding:"required,min=1"`
	Page      int32 `form:"page" binding:"required,min=1"`
	Size      int32 `form:"size" binding:"required,min=5,max=10"`
}

// listEntries returns paginated entries for an account owned by the authenticated user.
// @Summary List entries
// @Description List entries for a specific account with pagination.
// @Tags entries
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param account_id query int true "Account ID"
// @Param page query int true "Page number"
// @Param size query int true "Page size"
// @Success 200 {array} EntryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /entries [get]
func (server *Server) listEntries(ctx *gin.Context) {
	var req listEntriesRequest

	if err := ctx.ShouldBindWith(&req, binding.Query); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var (
		entries []db.Entry
		err     error
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

	if account.Owner != authPayload.Username {
		err := errors.New("entries don't belong to authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	arg := db.ListAccountEntriesParams{
		AccountID: req.AccountID,
		Limit:     req.Size,
		Offset:    (req.Page - 1) * req.Size,
	}
	entries, err = server.store.ListAccountEntries(ctx, arg)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, entries)
}
