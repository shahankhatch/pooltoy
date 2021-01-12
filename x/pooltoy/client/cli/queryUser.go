package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/interchainberlin/pooltoy/x/pooltoy/types"
	"github.com/spf13/cobra"
)

func GetCmdListUsers(queryRoute string, cdc *codec.AminoCodec) *cobra.Command {
	return &cobra.Command{
		Use:   "list-users",
		Short: "list all users",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := client.GetClientContextFromCmd(cmd)
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/"+types.QueryListUsers, queryRoute), nil)
			if err != nil {
				fmt.Printf("could not list User\n%s\n", err.Error())
				return nil
			}
			var out []types.User
			cliCtx.LegacyAmino.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintObjectLegacy(out)
		},
	}
}
