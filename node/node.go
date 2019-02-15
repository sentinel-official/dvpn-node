package node

import (
	"github.com/ironman0x7b2/vpn-node/server"
	"github.com/ironman0x7b2/vpn-node/tx"
	"github.com/ironman0x7b2/vpn-node/vpn"
)

type Node struct {
	tx     *tx.Tx
	vpn    *vpn.BaseVPN
	server *server.Server
}

func NewNode() *Node {
	return &Node{}
}
