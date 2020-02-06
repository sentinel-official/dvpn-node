module github.com/sentinel-official/dvpn-node

go 1.13

require (
	github.com/cosmos/cosmos-sdk v0.37.4
	github.com/cosmos/go-bip39 v0.0.0-20180819234021-555e2067c45d
	github.com/gorilla/mux v1.7.3
	github.com/gorilla/websocket v1.4.1
	github.com/jinzhu/gorm v1.9.11
	github.com/pelletier/go-toml v1.6.0
	github.com/pkg/errors v0.8.1
	github.com/sentinel-official/hub v0.2.0
	github.com/stretchr/testify v1.4.0
	github.com/tendermint/tendermint v0.32.8
)

replace github.com/sentinel-official/hub v0.2.0 => github.com/bitsndbyts/hub v0.2.1-0.20191226103641-58930d8aa800
