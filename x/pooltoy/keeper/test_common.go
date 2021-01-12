package keeper

import (
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	ibctransfertypes "github.com/cosmos/cosmos-sdk/x/ibc/applications/transfer/types"
	ibchost "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	//"github.com/cosmos/cosmos-sdk/x/supply"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	"github.com/cosmos/cosmos-sdk/x/bank"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	//"github.com/cosmos/cosmos-sdk/x/distribution"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	//"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	//"github.com/cosmos/cosmos-sdk/x/supply"
	"github.com/interchainberlin/pooltoy/x/pooltoy/types"
	"github.com/okwme/modules/incubator/faucet"

	//abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tendermintdbm "github.com/tendermint/tm-db"
)

var maccPerms = map[string][]string{
	authtypes.FeeCollectorName:     nil,
	disttypes.ModuleName:           nil,
	stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
	stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
	faucet.ModuleName:              {authtypes.Minter},
}

func CreateTestKeepers(t *testing.T) (Keeper, sdk.Context, bankkeeper.Keeper, authkeeper.AccountKeeper) {
	keyUser := sdk.NewKVStoreKey(types.StoreKey)
	keyAcc := sdk.NewKVStoreKey(authtypes.StoreKey)
	keyParams := sdk.NewKVStoreKey(paramstypes.StoreKey)
	tkeyParams := sdk.NewTransientStoreKey(paramstypes.TStoreKey)

	keys := sdk.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, stakingtypes.StoreKey,
		minttypes.StoreKey, disttypes.StoreKey, slashingtypes.StoreKey,
		govtypes.StoreKey, paramstypes.StoreKey, ibchost.StoreKey, upgradetypes.StoreKey,
		evidencetypes.StoreKey, ibctransfertypes.StoreKey, capabilitytypes.StoreKey,
	)

	db := tendermintdbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(keyUser, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyAcc, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)

	err := ms.LoadLatestVersion()
	require.Nil(t, err)

	ctx := sdk.NewContext(ms, tmproto.Header{ChainID: "testchainid"}, false, log.NewNopLogger())

	cdc := MakeTestCodec()
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	marshaler := codec.NewProtoCodec(interfaceRegistry)

	// module account permissions
	maccPerms = map[string][]string{
		authtypes.FeeCollectorName:     nil,
		disttypes.ModuleName:           nil,
		minttypes.ModuleName:           {authtypes.Minter},
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		govtypes.ModuleName:            {authtypes.Burner},
		ibctransfertypes.ModuleName:    {authtypes.Minter, authtypes.Burner},
	}

	paramsKeeper := paramskeeper.NewKeeper(marshaler, cdc.LegacyAmino, keyParams, tkeyParams)

	accountKeeper := authkeeper.NewAccountKeeper(marshaler, keys[authtypes.StoreKey], paramsKeeper.Subspace(authtypes.ModuleName), authtypes.ProtoBaseAccount, maccPerms)

	bankKeeper := bankkeeper.NewBaseKeeper(marshaler, keys[banktypes.StoreKey], accountKeeper, paramsKeeper.Subspace(banktypes.ModuleName), BlockedAccountAddrs())

	keeper := NewKeeper(bankKeeper, accountKeeper, marshaler, cdc.LegacyAmino, keyUser)

	return keeper, ctx, bankKeeper, accountKeeper
}

func BlockedAccountAddrs() map[string]bool {
	blockedAddrs := make(map[string]bool)
	for acc := range maccPerms {
		blockedAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	return blockedAddrs
}

func MakeTestCodec() *codec.AminoCodec {
	var cdc = codec.NewLegacyAmino()
	var ModuleCdc = codec.NewAminoCodec(cdc)
	auth.AppModuleBasic{}.RegisterLegacyAminoCodec(cdc)
	bank.AppModuleBasic{}.RegisterLegacyAminoCodec(cdc)
	params.AppModuleBasic{}.RegisterLegacyAminoCodec(cdc)
	vesting.AppModuleBasic{}.RegisterLegacyAminoCodec(cdc)
	types.RegisterCodec(ModuleCdc)
	sdk.RegisterLegacyAminoCodec(cdc)
	codec.RegisterEvidences(cdc)
	cryptocodec.RegisterCrypto(cdc)
	ModuleCdc.Seal()
	return ModuleCdc
}
