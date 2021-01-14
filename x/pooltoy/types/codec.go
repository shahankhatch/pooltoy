package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
)

// RegisterCodec registers concrete types on codec
func RegisterCodec(cdc *codec.AminoCodec) {
	// this line is used by starport scaffolding
	cdc.RegisterConcrete(&MsgCreateUser{}, "pooltoy/CreateUser", nil)
}

// ModuleCdc defines the module codec
var ModuleCdc *codec.AminoCodec
var Cdc *codec.LegacyAmino

func init() {
	Cdc = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(Cdc)
	RegisterCodec(ModuleCdc)
	cryptocodec.RegisterCrypto(Cdc)
	ModuleCdc.Seal()
}
