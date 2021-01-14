package app

import (
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	"github.com/cosmos/cosmos-sdk/simapp"
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
	pooltoyparams "github.com/interchainberlin/pooltoy/app/params"
	"github.com/interchainberlin/pooltoy/x/pooltoy"
	pooltoykeeper "github.com/interchainberlin/pooltoy/x/pooltoy/keeper"
	pooltoytypes "github.com/interchainberlin/pooltoy/x/pooltoy/types"
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

	config pooltoyparams.EncodingConfig

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
	pooltoyKeeper  pooltoykeeper.Keeper
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
	invCheckPeriod uint, pooltoyconfig pooltoyparams.EncodingConfig, baseAppOptions ...func(*bam.BaseApp),
) *PooltoyApp {

	bApp := bam.NewBaseApp(appName, logger, db, pooltoyconfig.TxConfig.TxDecoder(), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetAppVersion(version.Version)

	keys := sdk.NewKVStoreKeys(
		authtypes.StoreKey,
		stakingtypes.StoreKey,
		distrtypes.StoreKey,
		slashingtypes.StoreKey,
		paramstypes.StoreKey,
		pooltoytypes.StoreKey,
		faucet.StoreKey,
	)

	tKeys := sdk.NewTransientStoreKeys(
		paramstypes.TStoreKey,
	)

	var app = &PooltoyApp{
		BaseApp:        bApp,
		config:         pooltoyconfig,
		invCheckPeriod: invCheckPeriod,
		keys:           keys,
		tKeys:          tKeys,
		subspaces:      make(map[string]paramstypes.Subspace),
	}

	app.paramsKeeper = paramskeeper.NewKeeper(
		app.config.Marshaler,
		app.config.AminoCodec.LegacyAmino,
		keys[paramstypes.StoreKey],
		tKeys[paramstypes.TStoreKey],
	)

	app.subspaces[authtypes.ModuleName] = app.paramsKeeper.Subspace(authtypes.ModuleName)
	app.subspaces[banktypes.ModuleName] = app.paramsKeeper.Subspace(banktypes.ModuleName)
	app.subspaces[stakingtypes.ModuleName] = app.paramsKeeper.Subspace(stakingtypes.ModuleName)
	app.subspaces[distrtypes.ModuleName] = app.paramsKeeper.Subspace(distrtypes.ModuleName)
	app.subspaces[slashingtypes.ModuleName] = app.paramsKeeper.Subspace(slashingtypes.ModuleName)

	app.accountKeeper = authkeeper.NewAccountKeeper(
		app.config.Marshaler,
		keys[authtypes.StoreKey],
		app.subspaces[authtypes.ModuleName],
		authtypes.ProtoBaseAccount,
		maccPerms,
	)

	app.bankKeeper = bankkeeper.NewBaseKeeper(
		app.config.Marshaler,
		keys[banktypes.StoreKey],
		app.accountKeeper,
		app.subspaces[banktypes.ModuleName],
		app.BlockedAddrs(),
	)

	stakingKeeper := stakingkeeper.NewKeeper(
		app.config.Marshaler,
		keys[stakingtypes.StoreKey],
		app.accountKeeper,
		app.bankKeeper,
		app.subspaces[stakingtypes.ModuleName],
	)

	app.distrKeeper = distrkeeper.NewKeeper(
		app.config.Marshaler,
		keys[distrtypes.StoreKey],
		app.subspaces[distrtypes.ModuleName],
		app.accountKeeper,
		app.bankKeeper,
		&stakingKeeper,
		authtypes.FeeCollectorName,
		app.ModuleAccountAddrs(),
	)

	app.slashingKeeper = slashingkeeper.NewKeeper(
		app.config.Marshaler,
		keys[slashingtypes.StoreKey],
		&stakingKeeper,
		app.subspaces[slashingtypes.ModuleName],
	)

	app.stakingKeeper = *stakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(
			app.distrKeeper.Hooks(),
			app.slashingKeeper.Hooks()),
	)

	app.pooltoyKeeper = pooltoykeeper.NewKeeper(
		app.bankKeeper,
		app.accountKeeper,
		app.config.Marshaler,
		app.config.AminoCodec.LegacyAmino,
		keys[pooltoytypes.StoreKey],
		app.subspaces[slashingtypes.ModuleName],
	)

	app.faucetKeeper = faucet.NewKeeper(
		app.stakingKeeper,
		app.accountKeeper,
		1,            // amount for mint
		24*time.Hour, // rate limit by time
		keys[faucet.StoreKey],
		app.config.AminoCodec,
	)

	bankModule := bank.NewAppModule(app.config.Marshaler, app.bankKeeper, app.accountKeeper)
	restrictedBank := NewRestrictedBankModule(bankModule, app.bankKeeper, app.accountKeeper)

	app.mm = module.NewManager(
		genutil.NewAppModule(app.accountKeeper, app.stakingKeeper, app.BaseApp.DeliverTx, pooltoyconfig.TxConfig),
		auth.NewAppModule(app.config.Marshaler, app.accountKeeper, authsims.RandomGenesisAccounts),
		restrictedBank,
		distr.NewAppModule(app.config.Marshaler, app.distrKeeper, app.accountKeeper, app.bankKeeper, app.stakingKeeper),
		slashing.NewAppModule(app.config.Marshaler, app.slashingKeeper, app.accountKeeper, app.bankKeeper, app.stakingKeeper),
		pooltoy.NewAppModule(app.pooltoyKeeper, app.bankKeeper),
		faucet.NewAppModule(app.faucetKeeper),
		staking.NewAppModule(app.config.Marshaler, app.stakingKeeper, app.accountKeeper, app.bankKeeper),
		slashing.NewAppModule(app.config.Marshaler, app.slashingKeeper, app.accountKeeper, app.bankKeeper, app.stakingKeeper),
	)

	app.mm.SetOrderBeginBlockers(distrtypes.ModuleName, slashingtypes.ModuleName)
	app.mm.SetOrderEndBlockers(stakingtypes.ModuleName)

	app.mm.SetOrderInitGenesis(
		distrtypes.ModuleName,
		stakingtypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		slashingtypes.ModuleName,
		pooltoytypes.ModuleName,
		genutiltypes.ModuleName,
	)

	app.mm.RegisterRoutes(app.Router(), app.QueryRouter(), pooltoyconfig.AminoCodec.LegacyAmino)

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
	return app.mm.InitGenesis(ctx, app.config.Marshaler, genesisState)
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
	return app.config.AminoCodec
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

// LegacyAmino returns SimApp's amino codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *PooltoyApp) LegacyAmino() *codec.LegacyAmino {
	return app.config.AminoCodec.LegacyAmino
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
	authtx.RegisterTxService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.BaseApp.Simulate, app.config.InterfaceRegistry)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *PooltoyApp) RegisterTendermintService(clientCtx client.Context) {
	tmservice.RegisterTendermintService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.config.InterfaceRegistry)
}
