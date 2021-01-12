package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/interchainberlin/pooltoy/x/pooltoy/types"
)

func GetCmdCreateUser(cdc *codec.AminoCodec) *cobra.Command {
	return &cobra.Command{
		Use:   "create-user [userAccount] [isAdmin] [name] [email]",
		Short: "Creates a new user",
		Args:  cobra.MinimumNArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			argsUserAccount := string(args[0])
			userAccount, err := sdk.AccAddressFromBech32(argsUserAccount)
			if err != nil {
				return err
			}
			var isAdmin bool
			isAdmin, err = strconv.ParseBool(args[1])
			if err != nil {
				return err
			}
			argsName := string(args[2])
			argsEmail := string(args[3])

			clientCtx := client.GetClientContextFromCmd(cmd)
			msg := types.NewMsgCreateUser(clientCtx.GetFromAddress(), userAccount, isAdmin, argsName, argsEmail)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
}
