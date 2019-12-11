package cli

import (
	"io/ioutil"
	"os"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/lite"
	"github.com/tendermint/tendermint/lite/proxy"
	"github.com/tendermint/tendermint/rpc/client"

	"github.com/sentinel-official/dvpn-node/types"
)

type CLI struct {
	context.CLIContext
}

func NewCLI(cdc *codec.Codec, kb keys.Keybase, rpcAddress string, keyInfo keys.Info) *CLI {
	return &CLI{
		context.CLIContext{
			Codec:         cdc,
			Keybase:       kb,
			Output:        os.Stdout,
			OutputFormat:  "text",
			NodeURI:       rpcAddress,
			From:          keyInfo.GetName(),
			BroadcastMode: "sync",
			VerifierHome:  types.DefaultConfigDir,
			FromAddress:   keyInfo.GetAddress(),
			FromName:      keyInfo.GetName(),
			SkipConfirm:   true,
		},
	}
}

func NewVerifier(dir, id, address string) (*client.HTTP, *lite.DynamicVerifier, error) {
	root, err := ioutil.TempDir(dir, "lite_")
	if err != nil {
		return nil, nil, err
	}

	c := client.NewHTTP(address, "/websocket")

	verifier, err := proxy.NewVerifier(id, root, c, log.NewNopLogger(), 10)
	if err != nil {
		return nil, nil, err
	}

	return c, verifier, nil
}
