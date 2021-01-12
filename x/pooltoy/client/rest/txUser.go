package rest

import (
	//"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	//"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/interchainberlin/pooltoy/x/pooltoy/types"
)

type createUserRequest struct {
	BaseReq     rest.BaseReq `json:"base_req"`
	Creator     string       `json:"creator"`
	UserAccount string       `json:"userAccount"`
	IsAdmin     bool         `json:"isAdmin"`
	Name        string       `json:"name"`
	Email       string       `json:"email"`
}

func createUserHandler(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createUserRequest
		if !rest.ReadRESTReq(w, r, cliCtx.LegacyAmino, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}
		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}
		creator, err := sdk.AccAddressFromBech32(req.Creator)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		userAccount, err := sdk.AccAddressFromBech32(req.UserAccount)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		isAdmin := false

		msg := types.NewMsgCreateUser(creator, userAccount, isAdmin, req.Name, req.Email)
		tx.WriteGeneratedTxResponse(cliCtx, w, baseReq, msg)
	}
}
