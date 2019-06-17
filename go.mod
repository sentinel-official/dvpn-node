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
	github.com/prometheus/client_golang v1.0.0 // indirect
	github.com/spf13/cobra v0.0.5 // indirect
	github.com/stretchr/testify v1.3.0
	github.com/tendermint/tendermint v0.31.5
	golang.org/x/crypto v0.0.0-20190611184440-5c40567a22f8 // indirect
	golang.org/x/net v0.0.0-20190613194153-d28f0bde5980 // indirect
	golang.org/x/sys v0.0.0-20190616124812-15dcb6c0061f // indirect
	google.golang.org/appengine v1.6.1 // indirect
	google.golang.org/genproto v0.0.0-20190611190212-a7e196e89fd3 // indirect
	google.golang.org/grpc v1.21.1 // indirect
)

replace golang.org/x/crypto => github.com/tendermint/crypto v0.0.0-20180820045704-3764759f34a5
