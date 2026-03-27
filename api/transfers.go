package api

import (
	"database/sql"
	"net/http"

	db "github.com/HyperNaser/gobank/db/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type getTransferRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

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

	ctx.JSON(http.StatusOK, transfer)
}

type listAccountTransfersRequest struct {
	AccountID *int64 `form:"account_id" binding:"omitempty,min=1"`
	Page      int32  `form:"page" binding:"required,min=1"`
	Size      int32  `form:"size" binding:"required,min=5,max=10"`
}

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

	if req.AccountID != nil {
		arg := db.ListAccountTransfersParams{
			FromAccountID: *req.AccountID,
			Limit:         req.Size,
			Offset:        (req.Page - 1) * req.Size,
		}
		transfers, err = server.store.ListAccountTransfers(ctx, arg)
	} else {
		arg := db.ListTransfersParams{
			Limit:  req.Size,
			Offset: (req.Page - 1) * req.Size,
		}
		transfers, err = server.store.ListTransfers(ctx, arg)

	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, transfers)
}
