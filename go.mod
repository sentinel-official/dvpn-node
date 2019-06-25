module github.com/sentinel-official/dvpn-node

go 1.12

require (
	cloud.google.com/go v0.40.0 // indirect
	github.com/cosmos/cosmos-sdk v0.35.0
	github.com/cosmos/go-bip39 v0.0.0-20180819234021-555e2067c45d
	github.com/gorilla/mux v1.7.2
	github.com/gorilla/websocket v1.4.0
	github.com/jinzhu/gorm v1.9.9
	github.com/pelletier/go-toml v1.4.0
	github.com/pkg/errors v0.8.1
	github.com/sentinel-official/hub v0.0.0-20190625121651-69d3f09ffc37
	github.com/stretchr/testify v1.3.0
	github.com/tendermint/tendermint v0.31.5
	google.golang.org/appengine v1.6.1 // indirect
)

replace golang.org/x/crypto => github.com/tendermint/crypto v0.0.0-20180820045704-3764759f34a5
