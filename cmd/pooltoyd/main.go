package main

import (
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	"io"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	tmcli "github.com/tendermint/tendermint/libs/cli"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/debug"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/store"
	storeTypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	"github.com/interchainberlin/pooltoy/app"
	pooltoyparams "github.com/interchainberlin/pooltoy/params"
)

const flagInvCheckPeriod = "inv-check-period"

var invCheckPeriod uint

func main() {
	cdc := app.MakeEncodingConfig()

	newDnmRegex := `[\x{1F600}-\x{1F6FF}]`
	sdk.SetCoinDenomRegex(func() string {
		return newDnmRegex
	})

	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(sdk.Bech32PrefixAccAddr, sdk.Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(sdk.Bech32PrefixValAddr, sdk.Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(sdk.Bech32PrefixConsAddr, sdk.Bech32PrefixConsPub)
	config.Seal()

	ctx := server.NewDefaultContext()
	cobra.EnableCommandSorting = false
	rootCmd := &cobra.Command{
		Use:               "pooltoyd",
		Short:             "app Daemon (server)",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error { return nil },
	}

	rootCmd.AddCommand(genutilcli.InitCmd(app.ModuleBasics, app.DefaultNodeHome))
	rootCmd.AddCommand(genutilcli.CollectGenTxsCmd(banktypes.GenesisBalancesIterator{}, app.DefaultNodeHome))
	rootCmd.AddCommand(genutilcli.MigrateGenesisCmd())
	rootCmd.AddCommand(
		genutilcli.GenTxCmd(app.ModuleBasics, cdc.TxConfig, banktypes.GenesisBalancesIterator{}, app.DefaultNodeHome),
	)
	rootCmd.AddCommand(genutilcli.ValidateGenesisCmd(app.ModuleBasics))
	rootCmd.AddCommand(AddGenesisAccountCmd(ctx, cdc.AminoCodec, app.DefaultNodeHome, app.DefaultCLIHome))
	rootCmd.AddCommand(tmcli.NewCompletionCmd(rootCmd, true))
	rootCmd.AddCommand(debug.Cmd())

	encodingConfig := app.MakeEncodingConfig()
	a := appCreator{encodingConfig}
	server.AddCommands(rootCmd, app.DefaultNodeHome, a.newApp, a.appExport, addModuleInitFlags)

	// prepare and add flags
	executor := tmcli.PrepareBaseCmd(rootCmd, "AU", app.DefaultNodeHome)
	rootCmd.PersistentFlags().UintVar(&invCheckPeriod, flagInvCheckPeriod,
		0, "Assert registered invariants every N blocks")
	err := executor.Execute()
	if err != nil {
		panic(err)
	}
}

type appCreator struct {
	encCfg pooltoyparams.EncodingConfig
}

func (a appCreator) newApp(logger log.Logger, db dbm.DB, traceStore io.Writer, appOpts servertypes.AppOptions) servertypes.Application {
	var cache sdk.MultiStorePersistentCache

	if viper.GetBool(server.FlagInterBlockCache) {
		cache = store.NewCommitKVStoreCacheManager()
	}

	return app.NewPooltoyApp(
		logger, db, traceStore, true, invCheckPeriod,
		baseapp.SetPruning(storeTypes.NewPruningOptionsFromString(viper.GetString("pruning"))),
		baseapp.SetMinGasPrices(viper.GetString(server.FlagMinGasPrices)),
		baseapp.SetHaltHeight(viper.GetUint64(server.FlagHaltHeight)),
		baseapp.SetHaltTime(viper.GetUint64(server.FlagHaltTime)),
		baseapp.SetInterBlockCache(cache),
	)
}

func (a appCreator) appExport(
	logger log.Logger, db dbm.DB, traceStore io.Writer, height int64, forZeroHeight bool, jailWhiteList []string, appOpts servertypes.AppOptions) (servertypes.ExportedApp, error) {

	if height != -1 {
		aApp := app.NewPooltoyApp(logger, db, traceStore, false, uint(1))
		err := aApp.LoadHeight(height)
		if err != nil {
			return servertypes.ExportedApp{}, err
		}
		return aApp.ExportAppStateAndValidators(forZeroHeight, jailWhiteList)
	}

	aApp := app.NewPooltoyApp(logger, db, traceStore, true, uint(1))

	return aApp.ExportAppStateAndValidators(forZeroHeight, jailWhiteList)
}

func addModuleInitFlags(startCmd *cobra.Command) {
	crisis.AddModuleInitFlags(startCmd)
}
