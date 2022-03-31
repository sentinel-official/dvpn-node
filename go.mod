module github.com/sentinel-official/dvpn-node

go 1.16

require (
	github.com/avast/retry-go v3.0.0+incompatible
	github.com/cosmos/cosmos-sdk v0.44.6
	github.com/cosmos/go-bip39 v1.0.0
	github.com/go-kit/kit v0.12.0
	github.com/gorilla/mux v1.8.0
	github.com/pkg/errors v0.9.1
	github.com/rs/cors v1.8.2
	github.com/rs/zerolog v1.26.1
	github.com/sentinel-official/hub v0.9.0
	github.com/showwin/speedtest-go v1.1.5
	github.com/soheilhy/cmux v0.1.5
	github.com/spf13/cobra v1.4.0
	github.com/spf13/viper v1.10.1
	github.com/tendermint/tendermint v0.34.14
	golang.org/x/crypto v0.0.0-20211215165025-cf75a172585e
	google.golang.org/grpc v1.43.0
	gorm.io/driver/sqlite v1.3.1
	gorm.io/gorm v1.23.3
)

replace (
	github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
	google.golang.org/grpc => google.golang.org/grpc v1.33.2
)
