package app

import (
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authrest "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/gorilla/mux"
	"github.com/rakyll/statik/fs"
	tmjson "github.com/tendermint/tendermint/libs/json"
	"net/http"

	"io"
	"os"
	"time"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmos "github.com/tendermint/tendermint/libs/os"
	dbm "github.com/tendermint/tm-db"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	pooltoyparams "github.com/interchainberlin/pooltoy/params"
	"github.com/interchainberlin/pooltoy/x/pooltoy"
	"github.com/okwme/modules/incubator/faucet"
)

const appName = "pooltoy"

var (
	DefaultCLIHome  = os.ExpandEnv("$HOME/.pooltoycli")
	DefaultNodeHome = os.ExpandEnv("$HOME/.pooltoyd")
	ModuleBasics    = module.NewBasicManager(
		genutil.AppModuleBasic{},
		auth.AppModuleBasic{},
		bank.AppModuleBasic{},
		staking.AppModuleBasic{},
		distr.AppModuleBasic{},
		params.AppModuleBasic{},
		slashing.AppModuleBasic{},
		pooltoy.AppModuleBasic{},
		faucet.AppModule{},
	)

	maccPerms = map[string][]string{
		authtypes.FeeCollectorName:     nil,
		distrtypes.ModuleName:          nil,
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		faucet.ModuleName:              {authtypes.Minter},
	}

	// module accounts that are allowed to receive tokens
	allowedReceivingModAcc = map[string]bool{
		distrtypes.ModuleName: true,
	}
)

type PooltoyApp struct {
	*bam.BaseApp
	//cdc *codec.AminoCodec

	//legacyAmino *codec.LegacyAmino
	//appCodec    codec.Marshaler

	enc pooltoyparams.EncodingConfig

	interfaceRegistry types.InterfaceRegistry

	invCheckPeriod uint

	keys  map[string]*sdk.KVStoreKey
	tKeys map[string]*sdk.TransientStoreKey

	subspaces map[string]paramstypes.Subspace

	accountKeeper  authkeeper.AccountKeeper
	bankKeeper     bankkeeper.Keeper
	stakingKeeper  stakingkeeper.Keeper
	slashingKeeper slashingkeeper.Keeper
	distrKeeper    distrkeeper.Keeper
	paramsKeeper   paramskeeper.Keeper
	pooltoyKeeper  pooltoy.Keeper
	faucetKeeper   faucet.Keeper

	mm *module.Manager

	sm *module.SimulationManager
}

var (
	_ simapp.App       = (*PooltoyApp)(nil)
	_ abci.Application = (*PooltoyApp)(nil)
)

func NewPooltoyApp(
	logger log.Logger, db dbm.DB, traceStore io.Writer, loadLatest bool,
	invCheckPeriod uint, baseAppOptions ...func(*bam.BaseApp),
) *PooltoyApp {
	cdc := MakeEncodingConfig()

	pooltoyconfig := MakeEncodingConfig()
	bApp := bam.NewBaseApp(appName, logger, db, pooltoyconfig.TxConfig.TxDecoder(), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetAppVersion(version.Version)

	keys := sdk.NewKVStoreKeys(authtypes.StoreKey, stakingtypes.StoreKey,
		distrtypes.StoreKey, slashingtypes.StoreKey, paramstypes.StoreKey, pooltoy.StoreKey, faucet.StoreKey)

	tKeys := sdk.NewTransientStoreKeys(paramstypes.TStoreKey)

	var app = &PooltoyApp{
		BaseApp:        bApp,
		enc:            cdc,
		invCheckPeriod: invCheckPeriod,
		keys:           keys,
		tKeys:          tKeys,
		subspaces:      make(map[string]paramstypes.Subspace),
	}

	app.paramsKeeper = paramskeeper.NewKeeper(app.enc.Marshaler, app.enc.LegacyAmino, keys[paramstypes.StoreKey], tKeys[paramstypes.TStoreKey])
	app.subspaces[authtypes.ModuleName] = app.paramsKeeper.Subspace(authtypes.ModuleName)
	app.subspaces[banktypes.ModuleName] = app.paramsKeeper.Subspace(banktypes.ModuleName)
	app.subspaces[stakingtypes.ModuleName] = app.paramsKeeper.Subspace(stakingtypes.ModuleName)
	app.subspaces[distrtypes.ModuleName] = app.paramsKeeper.Subspace(distrtypes.ModuleName)
	app.subspaces[slashingtypes.ModuleName] = app.paramsKeeper.Subspace(slashingtypes.ModuleName)

	app.accountKeeper = authkeeper.NewAccountKeeper(
		app.enc.Marshaler,
		keys[authtypes.StoreKey],
		app.subspaces[authtypes.ModuleName],
		authtypes.ProtoBaseAccount,
		maccPerms,
	)

	app.bankKeeper = bankkeeper.NewBaseKeeper(
		app.enc.Marshaler,
		keys[banktypes.StoreKey],
		app.accountKeeper,
		app.subspaces[banktypes.ModuleName],
		app.BlockedAddrs(),
	)

	stakingKeeper := stakingkeeper.NewKeeper(
		app.enc.Marshaler,
		keys[stakingtypes.StoreKey],
		app.accountKeeper,
		app.bankKeeper,
		app.subspaces[stakingtypes.ModuleName],
	)

	app.distrKeeper = distrkeeper.NewKeeper(
		app.enc.Marshaler,
		keys[distrtypes.StoreKey],
		app.subspaces[distrtypes.ModuleName],
		app.accountKeeper,
		app.bankKeeper,
		&stakingKeeper,
		authtypes.FeeCollectorName,
		app.ModuleAccountAddrs(),
	)

	app.slashingKeeper = slashingkeeper.NewKeeper(
		app.enc.Marshaler,
		keys[slashingtypes.StoreKey],
		&stakingKeeper,
		app.subspaces[slashingtypes.ModuleName],
	)

	app.stakingKeeper = *stakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(
			app.distrKeeper.Hooks(),
			app.slashingKeeper.Hooks()),
	)

	app.pooltoyKeeper = pooltoy.NewKeeper(
		app.bankKeeper,
		app.accountKeeper,
		app.enc.Marshaler,
		app.enc.LegacyAmino,
		keys[pooltoy.StoreKey],
	)

	app.faucetKeeper = faucet.NewKeeper(
		app.stakingKeeper,
		app.accountKeeper,
		1,            // amount for mint
		24*time.Hour, // rate limit by time
		keys[faucet.StoreKey],
		app.enc.AminoCodec,
	)

	bankModule := bank.NewAppModule(app.enc.Marshaler, app.bankKeeper, app.accountKeeper)
	restrictedBank := NewRestrictedBankModule(bankModule, app.bankKeeper, app.accountKeeper)

	app.mm = module.NewManager(
		genutil.NewAppModule(app.accountKeeper, app.stakingKeeper, app.BaseApp.DeliverTx, pooltoyconfig.TxConfig),
		auth.NewAppModule(app.enc.Marshaler, app.accountKeeper, authsims.RandomGenesisAccounts),
		restrictedBank,
		distr.NewAppModule(app.enc.Marshaler, app.distrKeeper, app.accountKeeper, app.bankKeeper, app.stakingKeeper),
		slashing.NewAppModule(app.enc.Marshaler, app.slashingKeeper, app.accountKeeper, app.bankKeeper, app.stakingKeeper),
		pooltoy.NewAppModule(app.pooltoyKeeper, app.bankKeeper),
		faucet.NewAppModule(app.faucetKeeper),
		staking.NewAppModule(app.enc.Marshaler, app.stakingKeeper, app.accountKeeper, app.bankKeeper),
		slashing.NewAppModule(app.enc.Marshaler, app.slashingKeeper, app.accountKeeper, app.bankKeeper, app.stakingKeeper),
	)

	app.mm.SetOrderBeginBlockers(distrtypes.ModuleName, slashingtypes.ModuleName)
	app.mm.SetOrderEndBlockers(stakingtypes.ModuleName)

	app.mm.SetOrderInitGenesis(
		distrtypes.ModuleName,
		stakingtypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		slashingtypes.ModuleName,
		pooltoy.ModuleName,
		genutiltypes.ModuleName,
	)

	app.mm.RegisterRoutes(app.Router(), app.QueryRouter(), cdc.LegacyAmino)

	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetEndBlocker(app.EndBlocker)

	app.SetAnteHandler(
		ante.NewAnteHandler(
			app.accountKeeper, app.bankKeeper, ante.DefaultSigVerificationGasConsumer,
			pooltoyconfig.TxConfig.SignModeHandler(),
		),
	)

	app.MountKVStores(keys)
	app.MountTransientStores(tKeys)

	if loadLatest {
		err := app.LoadLatestVersion()
		if err != nil {
			tmos.Exit(err.Error())
		}
	}

	return app
}

type GenesisState map[string]json.RawMessage

//func NewDefaultGenesisState() GenesisState {
//	return ModuleBasics.DefaultGenesis()
//}

func (app *PooltoyApp) InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	//var genesisState simapp.GenesisState
	var genesisState GenesisState

	if err := tmjson.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}
	return app.mm.InitGenesis(ctx, app.enc.Marshaler, genesisState)
}

func (app *PooltoyApp) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	return app.mm.BeginBlock(ctx, req)
}

func (app *PooltoyApp) EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	return app.mm.EndBlock(ctx, req)
}

func (app *PooltoyApp) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

func (app *PooltoyApp) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

func (app *PooltoyApp) Codec() *codec.AminoCodec {
	return app.enc.AminoCodec
}

func (app *PooltoyApp) SimulationManager() *module.SimulationManager {
	return app.sm
}

// BlockedAddrs returns all the app's module account addresses that are not
// allowed to receive external tokens.
func (app *PooltoyApp) BlockedAddrs() map[string]bool {
	blockedAddrs := make(map[string]bool)
	for acc := range maccPerms {
		blockedAddrs[authtypes.NewModuleAddress(acc).String()] = !allowedReceivingModAcc[acc]
	}

	return blockedAddrs
}

func registerCodecsAndInterfaces(encodingConfig pooltoyparams.EncodingConfig) {
	std.RegisterLegacyAminoCodec(encodingConfig.LegacyAmino)
	std.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	sdk.RegisterLegacyAminoCodec(encodingConfig.LegacyAmino)
	sdk.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	codec.RegisterEvidences(encodingConfig.LegacyAmino)
	ModuleBasics.RegisterLegacyAminoCodec(encodingConfig.LegacyAmino)
	ModuleBasics.RegisterInterfaces(encodingConfig.InterfaceRegistry)
}

func MakeEncodingConfig() pooltoyparams.EncodingConfig {
	cdc := codec.NewLegacyAmino()
	interfaceRegistry := types.NewInterfaceRegistry()
	marshaler := codec.NewProtoCodec(interfaceRegistry)

	enc := pooltoyparams.EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Marshaler:         marshaler,
		TxConfig:          tx.NewTxConfig(marshaler, tx.DefaultSignModes),
		LegacyAmino:       cdc,
	}

	registerCodecsAndInterfaces(enc)

	return enc
}

// LegacyAmino returns SimApp's amino codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *PooltoyApp) LegacyAmino() *codec.LegacyAmino {
	return app.enc.LegacyAmino
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *PooltoyApp) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	clientCtx := apiSvr.ClientCtx
	rpc.RegisterRoutes(clientCtx, apiSvr.Router)
	// Register legacy tx routes.
	authrest.RegisterTxRoutes(clientCtx, apiSvr.Router)
	// Register new tx routes from grpc-gateway.
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	// Register new tendermint queries routes from grpc-gateway.
	tmservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register legacy and grpc-gateway routes for all modules.
	ModuleBasics.RegisterRESTRoutes(clientCtx, apiSvr.Router)
	ModuleBasics.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// register swagger API from root so that other applications can override easily
	if apiConfig.Swagger {
		RegisterSwaggerAPI(apiSvr.Router)
	}
}

// RegisterSwaggerAPI registers swagger route with API Server
func RegisterSwaggerAPI(rtr *mux.Router) {
	statikFS, err := fs.New()
	if err != nil {
		panic(err)
	}

	staticServer := http.FileServer(statikFS)
	rtr.PathPrefix("/swagger/").Handler(http.StripPrefix("/swagger/", staticServer))
}

// RegisterTxService implements the Application.RegisterTxService method.
func (app *PooltoyApp) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.BaseApp.Simulate, app.interfaceRegistry)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *PooltoyApp) RegisterTendermintService(clientCtx client.Context) {
	tmservice.RegisterTendermintService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.interfaceRegistry)
}
