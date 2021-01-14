package pooltoy

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	pooltoykeeper "github.com/interchainberlin/pooltoy/x/pooltoy/keeper"
	pooltoytypes "github.com/interchainberlin/pooltoy/x/pooltoy/types"
)

// InitGenesis initialize default parameters
// and the keeper's address to pubkey map
func InitGenesis(ctx sdk.Context, k pooltoykeeper.Keeper /* TODO: Define what keepers the module needs */, data pooltoytypes.GenesisState) {
	// TODO: Define logic for when you would like to initalize a new genesis
}

// ExportGenesis writes the current store values
// to a genesis file, which can be imported again
// with InitGenesis
func ExportGenesis(ctx sdk.Context, k pooltoykeeper.Keeper) (data pooltoytypes.GenesisState) {
	// TODO: Define logic for exporting state
	return pooltoytypes.NewGenesisState()
}
