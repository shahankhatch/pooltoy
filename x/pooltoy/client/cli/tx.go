package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/interchainberlin/pooltoy/x/pooltoy/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(cdc *codec.AminoCodec) *cobra.Command {
	pooltoyTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	createuserCmd := GetCmdCreateUser(cdc)
	flags.AddTxFlagsToCmd(createuserCmd)
	pooltoyTxCmd.AddCommand(createuserCmd)

	return pooltoyTxCmd
}
