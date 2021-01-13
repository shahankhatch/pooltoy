module github.com/interchainberlin/pooltoy

go 1.15

//replace github.com/cosmos/cosmos-sdk => /Users/shahank/git_interchain/cosmos-sdk

replace github.com/okwme/modules/incubator/faucet => /Users/shahank/git_interchain/faucet_modules/incubator/faucet
replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.2-alpha.regen.4

require (
	github.com/DataDog/zstd v1.4.5 // indirect
	github.com/cosmos/cosmos-sdk v0.40.0
	github.com/dgraph-io/badger/v2 v2.2007.2 // indirect
	github.com/dgraph-io/ristretto v0.0.3 // indirect
	github.com/dgryski/go-farm v0.0.0-20200201041132-a6ae2369ad13 // indirect
	github.com/enigmampc/btcutil v1.0.3-0.20200723161021-e2fb6adb2a25 // indirect
	github.com/golang/mock v1.4.4 // indirect
	github.com/google/uuid v1.1.2
	github.com/gorilla/handlers v1.5.1 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/magiconair/properties v1.8.4 // indirect
	github.com/okwme/modules/incubator/faucet v0.0.0-20200719150004-606b92fc6e9c
	github.com/pelletier/go-toml v1.8.0 // indirect
	github.com/prometheus/client_golang v1.8.0 // indirect
	github.com/rakyll/statik v0.1.7
	github.com/regen-network/cosmos-proto v0.3.0 // indirect
	github.com/spf13/cast v1.3.1 // indirect
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
