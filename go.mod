module github.com/interchainberlin/pooltoy

go 1.15

//replace github.com/cosmos/cosmos-sdk => /Users/shahank/git_interchain/cosmos-sdk

replace github.com/okwme/modules/incubator/faucet => /Users/shahank/git_interchain/faucet_modules/incubator/faucet

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.2-alpha.regen.4

require (
	github.com/cosmos/cosmos-sdk v0.40.0
	github.com/cosmos/gaia/v3 v3.0.0
	github.com/google/uuid v1.1.2
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/okwme/modules/incubator/faucet v0.0.0-20200719150004-606b92fc6e9c
	github.com/rakyll/statik v0.1.7
	github.com/spf13/cobra v1.1.1
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.6.1
	github.com/tendermint/go-amino v0.16.0
	github.com/tendermint/tendermint v0.34.1
	github.com/tendermint/tm-db v0.6.3
)

// replace github.com/okwme/modules/incubator/faucet => /Users/billy/GitHub.com/okwme/modules/incubator/faucet

// replace github.com/cosmos/cosmos-sdk v0.38.4 => github.com/okwme/cosmos-sdk v0.38.6-0.20200802130156-46d1ad2d6210

// replace github.com/cosmos/cosmos-sdk v0.38.4 => /Users/billy/GitHub/cosmos/cosmos-sdk
