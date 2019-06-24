module github.com/ironman0x7b2/vpn-node

go 1.12

require (
	cloud.google.com/go v0.40.0 // indirect
	github.com/btcsuite/btcd v0.0.0-20190614013741-962a206e94e9 // indirect
	github.com/cosmos/cosmos-sdk v0.35.0
	github.com/cosmos/go-bip39 v0.0.0-20180819234021-555e2067c45d
	github.com/gorilla/mux v1.7.2
	github.com/gorilla/websocket v1.4.0
	github.com/ironman0x7b2/sentinel-sdk v0.0.0-20190613115318-ab50711cb791
	github.com/jinzhu/gorm v1.9.9
	github.com/pelletier/go-toml v1.4.0
	github.com/pkg/errors v0.8.1
	github.com/prometheus/common v0.6.0 // indirect
	github.com/spf13/cobra v0.0.5 // indirect
	github.com/stretchr/testify v1.3.0
	github.com/tendermint/tendermint v0.31.5
	golang.org/x/crypto v0.0.0-20190621222207-cc06ce4a13d4 // indirect
	golang.org/x/net v0.0.0-20190620200207-3b0461eec859 // indirect
	golang.org/x/sys v0.0.0-20190624142023-c5567b49c5d0 // indirect
	google.golang.org/appengine v1.6.1 // indirect
	google.golang.org/genproto v0.0.0-20190620144150-6af8c5fc6601 // indirect
	google.golang.org/grpc v1.21.1 // indirect
)

replace golang.org/x/crypto => github.com/tendermint/crypto v0.0.0-20180820045704-3764759f34a5
