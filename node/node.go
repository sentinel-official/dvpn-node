package node

import (
	"path"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/viper"

	"github.com/sentinel-official/dvpn-node/context"
	httputils "github.com/sentinel-official/dvpn-node/utils/http"
)

type Node struct {
	*context.Context
}

func NewNode(ctx *context.Context) *Node {
	return &Node{ctx}
}

func (n *Node) Initialize() error {
	n.Log().Info("Initializing...")

	result, err := n.Client().QueryNode(n.Address())
	if err != nil {
		return err
	}

	if result == nil {
		return n.RegisterNode()
	}

	return n.UpdateNodeInfo()
}

func (n *Node) Start() error {
	n.Log().Info("Starting...")

	go func() {
		if err := n.jobSetSessions(); err != nil {
			panic(err)
		}
	}()

	go func() {
		if err := n.jobUpdateStatus(); err != nil {
			panic(err)
		}
	}()

	go func() {
		if err := n.jobUpdateSessions(); err != nil {
			panic(err)
		}
	}()

	var (
		certFile = path.Join(viper.GetString(flags.FlagHome), "tls.crt")
		keyFile  = path.Join(viper.GetString(flags.FlagHome), "tls.key")
	)

	return httputils.ListenAndServeTLS(
		n.ListenOn(),
		certFile,
		keyFile,
		n.Handler(),
	)
}
