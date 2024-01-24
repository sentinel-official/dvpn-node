package node

import (
	"path"

	"github.com/sentinel-official/dvpn-node/context"
	"github.com/sentinel-official/dvpn-node/utils"
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

func (n *Node) Start(home string) error {
	go func() {
		if err := n.jobSetSessions(); err != nil {
			panic(err)
		}
	}()

	go func() {
		if err := n.jobUpdateSessions(); err != nil {
			panic(err)
		}
	}()

	go func() {
		if err := n.jobUpdateStatus(); err != nil {
			panic(err)
		}
	}()

	var (
		certFile = path.Join(home, "tls.crt")
		keyFile  = path.Join(home, "tls.key")
	)

	return utils.ListenAndServeTLS(
		n.ListenOn(),
		certFile,
		keyFile,
		n.Handler(),
	)
}
