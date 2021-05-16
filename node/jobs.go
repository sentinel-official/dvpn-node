package node

import (
	"encoding/base64"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	hubtypes "github.com/sentinel-official/hub/types"

	"github.com/sentinel-official/dvpn-node/types"
)

func (n *Node) jobUpdateStatus() error {
	n.Logger().Info("Started job", "name", "update_status", "interval", n.IntervalStatus())

	t := time.NewTicker(n.IntervalStatus())
	for ; ; <-t.C {
		if err := n.updateStatus(); err != nil {
			return err
		}
	}
}

func (n *Node) jobUpdateSessions() error {
	n.Logger().Info("Started job", "name", "update_sessions", "interval", n.IntervalSessions())

	t := time.NewTicker(n.IntervalSessions())
	for ; ; <-t.C {
		var (
			items []types.Session
		)

		peers, err := n.Service().Peers()
		if err != nil {
			return err
		}

		for _, peer := range peers {
			var (
				item     = n.Sessions().Get(peer.Key)
				consumed = sdk.NewInt(peer.Upload + peer.Download)
			)

			session, err := n.Client().QuerySession(item.ID)
			if err != nil {
				return err
			}

			if session.Status.Equal(hubtypes.Inactive) || consumed.GT(item.Available) {
				key, err := base64.StdEncoding.DecodeString(peer.Key)
				if err != nil {
					return err
				}

				if err := n.Service().RemovePeer(key); err != nil {
					return err
				}

				n.Sessions().Delete(peer.Key)
			}

			if session.Status.Equal(hubtypes.Inactive) {
				continue
			}

			item.Upload = peer.Upload
			item.Download = peer.Download
			items = append(items, item)
		}

		if err := n.updateSessions(items); err != nil {
			return err
		}
	}
}
