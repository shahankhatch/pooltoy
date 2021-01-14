package cli

import (
	"fmt"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/interchainberlin/pooltoy/x/pooltoy/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string, cdc *codec.AminoCodec) *cobra.Command {
	// Group pooltoy queries under a subcommand
	pooltoyQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmdListUsers := GetCmdListUsers(queryRoute, cdc)
	flags.AddTxFlagsToCmd(cmdListUsers)
	pooltoyQueryCmd.AddCommand(cmdListUsers)

	return pooltoyQueryCmd
}
