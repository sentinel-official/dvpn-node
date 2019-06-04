module github.com/ironman0x7b2/vpn-node

go 1.12

require (
	cloud.google.com/go v0.39.0 // indirect
	github.com/cosmos/cosmos-sdk v0.35.0
	github.com/cosmos/go-bip39 v0.0.0-20180819234021-555e2067c45d
	github.com/denisenkom/go-mssqldb v0.0.0-20190515213511-eb9f6a1743f3 // indirect
	github.com/google/go-cmp v0.3.0 // indirect
	github.com/gorilla/mux v1.7.2
	github.com/gorilla/websocket v1.4.0
	github.com/ironman0x7b2/sentinel-sdk v0.0.0-20190604122145-b342c8579555
	github.com/jinzhu/gorm v1.9.8
	github.com/jinzhu/inflection v0.0.0-20190603042836-f5c5f50e6090 // indirect
	github.com/lib/pq v1.1.1 // indirect
	github.com/pelletier/go-toml v1.4.0
	github.com/pkg/errors v0.8.1
	github.com/tendermint/tendermint v0.31.5
	google.golang.org/appengine v1.6.0 // indirect
)

replace golang.org/x/crypto => github.com/tendermint/crypto v0.0.0-20180820045704-3764759f34a5
