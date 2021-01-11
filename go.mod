module github.com/sentinel-official/dvpn-node

go 1.15

require (
	github.com/cosmos/cosmos-sdk v0.39.2
	github.com/cosmos/go-bip39 v0.0.0-20180819234021-555e2067c45d
	github.com/gorilla/mux v1.8.0
	github.com/pelletier/go-toml v1.8.0
	github.com/sentinel-official/hub v0.4.0-rc1
	github.com/spf13/cobra v1.0.0
	github.com/spf13/viper v1.7.1
	github.com/tendermint/tendermint v0.33.9
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	golang.org/x/sync v0.0.0-20201008141435-b3e1573b7520
)

replace (
	github.com/keybase/go-keychain => github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4
)
