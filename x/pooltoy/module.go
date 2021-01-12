package pooltoy

import (
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/client"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	//"github.com/cosmos/cosmos-sdk/client/context"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"

	abci "github.com/tendermint/tendermint/abci/types"

	//"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	//"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/interchainberlin/pooltoy/x/pooltoy/client/cli"
	"github.com/interchainberlin/pooltoy/x/pooltoy/client/rest"
)

// Type check to ensure the interface is properly implemented
var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// AppModuleBasic defines the basic application module used by the pooltoy module.
type AppModuleBasic struct{}

// Name returns the pooltoy module's name.
func (AppModuleBasic) Name() string {
	return ModuleName
}

// RegisterCodec registers the pooltoy module's types for the given codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	RegisterCodec(codec.NewAminoCodec(cdc))
}

// DefaultGenesis returns default genesis state as raw bytes for the pooltoy
// module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONMarshaler) json.RawMessage {
	return ModuleCdc.LegacyAmino.MustMarshalJSON(DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the pooltoy module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONMarshaler, config client.TxEncodingConfig, bz json.RawMessage) error {
	var data GenesisState
	err := ModuleCdc.LegacyAmino.UnmarshalJSON(bz, &data)
	if err != nil {
		return err
	}
	return ValidateGenesis(data)
}

// RegisterRESTRoutes registers the REST routes for the pooltoy module.
func (AppModuleBasic) RegisterRESTRoutes(ctx client.Context, rtr *mux.Router) {
	rest.RegisterRoutes(ctx, rtr)
}

// GetTxCmd returns the root tx command for the pooltoy module.
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.GetTxCmd(ModuleCdc)
}

// GetQueryCmd returns no root query command for the pooltoy module.
func (a AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd(StoreKey, ModuleCdc)
}

//____________________________________________________________________________

// AppModule implements an application module for the pooltoy module.
type AppModule struct {
	AppModuleBasic

	keeper     Keeper
	coinKeeper bankkeeper.Keeper
	// TODO: Add keepers that your application depends on

}

// NewAppModule creates a new AppModule object
func NewAppModule(k Keeper, bankKeeper bankkeeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper:         k,
		coinKeeper:     bankKeeper,
		// TODO: Add keepers that your application depends on
	}
}

// Name returns the pooltoy module's name.
func (AppModule) Name() string {
	return ModuleName
}

// RegisterInvariants registers the pooltoy module invariants.
func (am AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// Route returns the message routing key for the pooltoy module.
//func (AppModule) Route() string {
//	return RouterKey
//}

// Route returns the message routing key for the pooltoy module.
func (AppModule) Route() sdk.Route { return sdk.Route{} }

// NewHandler returns an sdk.Handler for the pooltoy module.
func (am AppModule) NewHandler() sdk.Handler {
	return NewHandler(am.keeper)
}

// QuerierRoute returns the pooltoy module's querier route name.
func (AppModule) QuerierRoute() string {
	return QuerierRoute
}

// NewQuerierHandler returns the pooltoy module sdk.Querier.
func (am AppModule) NewQuerierHandler() sdk.Querier {
	return NewQuerier(am.keeper)
}

// InitGenesis performs genesis initialization for the pooltoy module. It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONMarshaler, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState GenesisState
	ModuleCdc.LegacyAmino.MustUnmarshalJSON(data, &genesisState)
	InitGenesis(ctx, am.keeper, genesisState)
	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the exported genesis state as raw bytes for the pooltoy
// module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONMarshaler) json.RawMessage {
	gs := ExportGenesis(ctx, am.keeper)
	return ModuleCdc.LegacyAmino.MustMarshalJSON(gs)
}

// BeginBlock returns the begin blocker for the pooltoy module.
func (am AppModule) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) {
	BeginBlocker(ctx, req, am.keeper)
}

// EndBlock returns the end blocker for the pooltoy module. It returns no validator
// updates.
func (AppModule) EndBlock(_ sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

// RegisterInterfaces registers the module's interface types
func (b AppModuleBasic) RegisterInterfaces(_ codectypes.InterfaceRegistry) {}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the mint module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	//types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx))

}

// LegacyQuerierHandler returns the mint module sdk.Querier.
func (am AppModule) LegacyQuerierHandler(legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return am.NewQuerierHandler()
}

// RegisterServices registers a gRPC query service to respond to the
// module-specific gRPC queries.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	//types.RegisterQueryServer(cfg.QueryServer(), am.keeper)
}
