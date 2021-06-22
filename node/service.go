package node

import (
	"encoding/base64"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sentinel-official/dvpn-node/types"
)

func (n *Node) RemovePeer(key string) error {
	n.Log().Info("Removing peer from underlying service", "key", key)

	data, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return err
	}

	if err := n.Service().RemovePeer(data); err != nil {
		return err
	}

	n.Log().Debug("Removed peer from underlying service...")
	return nil
}

func (n *Node) RemoveSession(key string, address sdk.AccAddress) error {
	n.Log().Info("Removing session", "key", key, "address", address)

	n.Sessions().DeleteByKey(key)
	n.Sessions().DeleteByAddress(address)

	n.Log().Debug("Removed session...")
	return nil
}

func (n *Node) RemovePeerAndSession(v types.Session) error {
	n.Log().Info("Removing peer and session", "id", v.ID, "key", v.Key)

	if err := n.RemovePeer(v.Key); err != nil {
		return err
	}
	if err := n.RemoveSession(v.Key, v.Address); err != nil {
		return err
	}

	n.Log().Debug("Removed peer and session...")
	return nil
}
