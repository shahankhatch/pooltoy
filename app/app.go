package app

import (
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
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

	//"github.com/cosmos/cosmos-sdk/x/supply"
	"io"
	"os"
	"time"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmos "github.com/tendermint/tendermint/libs/os"
	dbm "github.com/tendermint/tm-db"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
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
	//"github.com/cosmos/cosmos-sdk/x/supply"
	"github.com/interchainberlin/pooltoy/x/pooltoy"
	"github.com/okwme/modules/incubator/faucet"
)

const appName = "app"

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
		//supply.AppModuleBasic{},
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

func MakeCodec() *codec.AminoCodec {
	var cdc = codec.NewLegacyAmino()

	ModuleBasics.RegisterLegacyAminoCodec(cdc)
	sdk.RegisterLegacyAminoCodec(cdc)
	codec.RegisterEvidences(cdc)
	cdc.Seal()
	return codec.NewAminoCodec(cdc)
}

type NewApp struct {
	*bam.BaseApp
	cdc *codec.AminoCodec

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
	//supplyKeeper   supply.Keeper
	paramsKeeper  paramskeeper.Keeper
	pooltoyKeeper pooltoy.Keeper
	faucetKeeper  faucet.Keeper

	mm *module.Manager

	sm *module.SimulationManager
}

var _ simapp.App = (*NewApp)(nil)

func NewInitApp(
	logger log.Logger, db dbm.DB, traceStore io.Writer, loadLatest bool,
	invCheckPeriod uint, baseAppOptions ...func(*bam.BaseApp),
) *NewApp {
	cdc := MakeCodec()

	config := MakeTestEncodingConfig()
	// authtx.DefaultTxDecoder(codec.NewProtoCodec(codectypes.NewInterfaceRegistry()))
	bApp := bam.NewBaseApp(appName, logger, db, config.TxConfig.TxDecoder(), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetAppVersion(version.Version)

	keys := sdk.NewKVStoreKeys(authtypes.StoreKey, stakingtypes.StoreKey,
		distrtypes.StoreKey, slashingtypes.StoreKey, paramstypes.StoreKey, pooltoy.StoreKey, faucet.StoreKey)

	tKeys := sdk.NewTransientStoreKeys(paramstypes.TStoreKey)

	var app = &NewApp{
		BaseApp:        bApp,
		cdc:            cdc,
		invCheckPeriod: invCheckPeriod,
		keys:           keys,
		tKeys:          tKeys,
		subspaces:      make(map[string]paramstypes.Subspace),
	}

	app.paramsKeeper = paramskeeper.NewKeeper(app.cdc, cdc.LegacyAmino, keys[paramstypes.StoreKey], tKeys[paramstypes.TStoreKey])
	app.subspaces[authtypes.ModuleName] = app.paramsKeeper.Subspace(authtypes.ModuleName)
	app.subspaces[banktypes.ModuleName] = app.paramsKeeper.Subspace(banktypes.ModuleName)
	app.subspaces[stakingtypes.ModuleName] = app.paramsKeeper.Subspace(stakingtypes.ModuleName)
	app.subspaces[distrtypes.ModuleName] = app.paramsKeeper.Subspace(distrtypes.ModuleName)
	app.subspaces[slashingtypes.ModuleName] = app.paramsKeeper.Subspace(slashingtypes.ModuleName)

	app.accountKeeper = authkeeper.NewAccountKeeper(
		app.cdc,
		keys[authtypes.StoreKey],
		app.subspaces[authtypes.ModuleName],
		authtypes.ProtoBaseAccount,
		maccPerms,
	)

	app.bankKeeper = bankkeeper.NewBaseKeeper(
		cdc,
		keys[banktypes.StoreKey],
		app.accountKeeper,
		app.subspaces[banktypes.ModuleName],
		app.BlockedAddrs(),
	)

	//app.supplyKeeper = supplykeeper.NewKeeper(
	//	app.cdc,
	//	keys[supply.StoreKey],
	//	app.accountKeeper,
	//	app.bankKeeper,
	//	maccPerms,
	//)

	stakingKeeper := stakingkeeper.NewKeeper(
		app.cdc,
		keys[stakingtypes.StoreKey],
		app.accountKeeper,
		app.bankKeeper,
		app.subspaces[stakingtypes.ModuleName],
	)

	app.distrKeeper = distrkeeper.NewKeeper(
		app.cdc,
		keys[distrtypes.StoreKey],
		app.subspaces[distrtypes.ModuleName],
		app.accountKeeper,
		app.bankKeeper,
		&stakingKeeper,
		authtypes.FeeCollectorName,
		app.ModuleAccountAddrs(),
	)

	app.slashingKeeper = slashingkeeper.NewKeeper(
		app.cdc,
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
		app.cdc,
		app.cdc.LegacyAmino,
		keys[pooltoy.StoreKey],
	)

	app.faucetKeeper = faucet.NewKeeper(
		app.stakingKeeper,
		app.accountKeeper,
		1,            // amount for mint
		24*time.Hour, // rate limit by time
		keys[faucet.StoreKey],
		app.cdc,
	)

	bankModule := bank.NewAppModule(cdc, app.bankKeeper, app.accountKeeper)
	restrictedBank := NewRestrictedBankModule(bankModule, app.bankKeeper, app.accountKeeper)

	app.mm = module.NewManager(
		genutil.NewAppModule(app.accountKeeper, app.stakingKeeper, app.BaseApp.DeliverTx, config.TxConfig),
		auth.NewAppModule(cdc, app.accountKeeper, authsims.RandomGenesisAccounts),
		// bank.NewAppModule(app.bankKeeper, app.accountKeeper),
		restrictedBank,
		distr.NewAppModule(cdc, app.distrKeeper, app.accountKeeper, app.bankKeeper, app.stakingKeeper),
		slashing.NewAppModule(cdc, app.slashingKeeper, app.accountKeeper, app.bankKeeper, app.stakingKeeper),
		pooltoy.NewAppModule(app.pooltoyKeeper, app.bankKeeper),
		faucet.NewAppModule(app.faucetKeeper),
		staking.NewAppModule(cdc, app.stakingKeeper, app.accountKeeper, app.bankKeeper),
		slashing.NewAppModule(cdc, app.slashingKeeper, app.accountKeeper, app.bankKeeper, app.stakingKeeper),
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
			config.TxConfig.SignModeHandler(),
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

func (app *NewApp) InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	//var genesisState simapp.GenesisState
	var genesisState GenesisState

	if err := tmjson.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}
	return app.mm.InitGenesis(ctx, app.cdc, genesisState)
}

func (app *NewApp) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	return app.mm.BeginBlock(ctx, req)
}

func (app *NewApp) EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	return app.mm.EndBlock(ctx, req)
}

func (app *NewApp) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

func (app *NewApp) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

func (app *NewApp) Codec() *codec.AminoCodec {
	return app.cdc
}

func (app *NewApp) SimulationManager() *module.SimulationManager {
	return app.sm
}

func GetMaccPerms() map[string][]string {
	modAccPerms := make(map[string][]string)
	for k, v := range maccPerms {
		modAccPerms[k] = v
	}
	return modAccPerms
}

// BlockedAddrs returns all the app's module account addresses that are not
// allowed to receive external tokens.
func (app *NewApp) BlockedAddrs() map[string]bool {
	blockedAddrs := make(map[string]bool)
	for acc := range maccPerms {
		blockedAddrs[authtypes.NewModuleAddress(acc).String()] = !allowedReceivingModAcc[acc]
	}

	return blockedAddrs
}

// MakeTestEncodingConfig creates an EncodingConfig for testing.
// This function should be used only internally (in the SDK).
// App user should'nt create new codecs - use the app.AppCodec instead.
// [DEPRECATED]
func MakeTestEncodingConfig() simappparams.EncodingConfig {
	encodingConfig := simappparams.MakeTestEncodingConfig()
	std.RegisterLegacyAminoCodec(encodingConfig.Amino)
	std.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	ModuleBasics.RegisterLegacyAminoCodec(encodingConfig.Amino)
	ModuleBasics.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	return encodingConfig
}

// LegacyAmino returns SimApp's amino codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *NewApp) LegacyAmino() *codec.LegacyAmino {
	return app.cdc.LegacyAmino
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *NewApp) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
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
		RegisterSwaggerAPI(clientCtx, apiSvr.Router)
	}
}

// RegisterSwaggerAPI registers swagger route with API Server
func RegisterSwaggerAPI(ctx client.Context, rtr *mux.Router) {
	statikFS, err := fs.New()
	if err != nil {
		panic(err)
	}

	staticServer := http.FileServer(statikFS)
	rtr.PathPrefix("/swagger/").Handler(http.StripPrefix("/swagger/", staticServer))
}

// RegisterTxService implements the Application.RegisterTxService method.
func (app *NewApp) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.BaseApp.Simulate, app.interfaceRegistry)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *NewApp) RegisterTendermintService(clientCtx client.Context) {
	tmservice.RegisterTendermintService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.interfaceRegistry)
}
