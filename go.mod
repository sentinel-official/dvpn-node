module github.com/sentinel-official/dvpn-node

go 1.16

require (
	github.com/avast/retry-go v3.0.0+incompatible
	github.com/cosmos/cosmos-sdk v0.42.5
	github.com/cosmos/go-bip39 v1.0.0
	github.com/go-kit/kit v0.11.0
	github.com/gorilla/mux v1.8.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.0
	github.com/rs/cors v1.8.0
	github.com/rs/zerolog v1.23.0
	github.com/sentinel-official/hub v0.7.0
	github.com/showwin/speedtest-go v1.1.2
	github.com/soheilhy/cmux v0.1.5
	github.com/spf13/cobra v1.2.1
	github.com/spf13/viper v1.8.1
	github.com/tendermint/tendermint v0.34.11
	golang.org/x/crypto v0.0.0-20210711020723-a769d52b0f97
	google.golang.org/grpc v1.39.0
	gorm.io/driver/sqlite v1.1.4
	gorm.io/gorm v1.21.12
)

replace (
	github.com/99designs/keyring => github.com/99designs/keyring v1.1.7-0.20210324095724-d9b6b92e219f
	github.com/cosmos/cosmos-sdk => github.com/sentinel-official/cosmos-sdk v0.42.6-sentinel
	github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
	github.com/showwin/speedtest-go => github.com/bsrinivas8687/speedtest-go v1.1.3-0.20210629110436-a23d0944a09d
	google.golang.org/grpc => google.golang.org/grpc v1.33.2
)
