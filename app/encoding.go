package app

import (
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/interchainberlin/pooltoy/app/params"
)

// MakeEncodingConfig creates an EncodingConfig for testing
func MakeEncodingConfig() params.EncodingConfig {
	encodingConfig := params.MakeConfig()
	std.RegisterLegacyAminoCodec(encodingConfig.AminoCodec.LegacyAmino)
	std.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	ModuleBasics.RegisterLegacyAminoCodec(encodingConfig.AminoCodec.LegacyAmino)
	ModuleBasics.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	return encodingConfig
}
