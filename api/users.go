package api

import (
	"net/http"
	"time"

	db "github.com/HyperNaser/gobank/db/sqlc"
	"github.com/HyperNaser/gobank/util"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/lib/pq"
)

type createUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=8"`
	FullName string `json:"full_name" binding:"required,alphaspace"`
	Email    string `json:"email" binding:"required,email"`
}

type createUserResponse struct {
	Username          string    `json:"username"`
	FullName          string    `json:"full_name"`
	Email             string    `json:"email"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at"`
}

func (server *Server) createUser(ctx *gin.Context) {
	var req createUserRequest
	if err := ctx.ShouldBindWith(&req, binding.JSON); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	arg := db.CreateUserParams{
		Username:       req.Username,
		HashedPassword: hashedPassword,
		FullName:       req.FullName,
		Email:          req.Email,
	}

	user, err := server.store.CreateUser(ctx, arg)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	resp := createUserResponse{
		Username:          user.Username,
		FullName:          user.FullName,
		Email:             user.Email,
		PasswordChangedAt: user.PasswordChangedAt,
		CreatedAt:         user.CreatedAt,
	}

	ctx.JSON(http.StatusOK, resp)
}

// type getAccountRequest struct {
// 	ID int64 `uri:"id" binding:"required,min=1"`
// }
//
// func (server *Server) getAccount(ctx *gin.Context) {
// 	var req getAccountRequest
// 	if err := ctx.ShouldBindUri(&req); err != nil {
// 		ctx.JSON(http.StatusBadRequest, errorResponse(err))
// 		return
// 	}
//
// 	account, err := server.store.GetAccount(ctx, req.ID)
// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			ctx.JSON(http.StatusNotFound, errorResponse(err))
// 			return
// 		}
// 		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
// 		return
// 	}
//
// 	ctx.JSON(http.StatusOK, account)
// }
//
// type deleteAccountRequest struct {
// 	ID int64 `uri:"id" binding:"required,min=1"`
// }
//
// func (server *Server) deleteAccount(ctx *gin.Context) {
// 	var req deleteAccountRequest
// 	if err := ctx.ShouldBindUri(&req); err != nil {
// 		ctx.JSON(http.StatusBadRequest, errorResponse(err))
// 		return
// 	}
//
// 	err := server.store.DeleteAccount(ctx, req.ID)
// 	if err != nil {
// 		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
// 		return
// 	}
//
// 	ctx.Status(http.StatusNoContent)
// }
//
// type listAccountsRequest struct {
// 	Page int32 `form:"page" binding:"required,min=1"`
// 	Size int32 `form:"size" binding:"required,min=5,max=10"`
// }
//
// func (server *Server) listAccounts(ctx *gin.Context) {
// 	var req listAccountsRequest
// 	if err := ctx.ShouldBindWith(&req, binding.Query); err != nil {
// 		ctx.JSON(http.StatusBadRequest, errorResponse(err))
// 		return
// 	}
//
// 	arg := db.ListAccountsParams{
// 		Limit:  req.Size,
// 		Offset: (req.Page - 1) * req.Size,
// 	}
//
// 	accounts, err := server.store.ListAccounts(ctx, arg)
// 	if err != nil {
// 		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
// 		return
// 	}
//
// 	ctx.JSON(http.StatusOK, accounts)
// }
