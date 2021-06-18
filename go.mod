module github.com/sentinel-official/dvpn-node

go 1.16

require (
	github.com/avast/retry-go v3.0.0+incompatible
	github.com/cosmos/cosmos-sdk v0.42.5
	github.com/cosmos/go-bip39 v1.0.0
	github.com/gorilla/mux v1.8.0
	github.com/itchyny/gojq v0.12.4
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.21.0
	github.com/sentinel-official/hub v0.6.3
	github.com/showwin/speedtest-go v1.1.2
	github.com/spf13/cobra v1.1.3
	github.com/spf13/viper v1.7.1
	github.com/tendermint/tendermint v0.34.10
	github.com/v2fly/v2ray-core/v4 v4.40.0
	golang.org/x/crypto v0.0.0-20210513164829-c07d793c2f9a
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	google.golang.org/grpc v1.38.0
)

replace (
	github.com/cosmos/cosmos-sdk => github.com/sentinel-official/cosmos-sdk v0.42.6-sentinel
	github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
	google.golang.org/grpc => google.golang.org/grpc v1.33.2
)
