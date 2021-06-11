package node

import (
	"encoding/base64"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	hubtypes "github.com/sentinel-official/hub/types"

	"github.com/sentinel-official/dvpn-node/types"
)

func (n *Node) jobUpdateStatus() error {
	n.Log().Info("Starting job", "name", "update_status", "interval", n.IntervalStatus())

	t := time.NewTicker(n.IntervalStatus())
	for ; ; <-t.C {
		if err := n.updateStatus(); err != nil {
			return err
		}
	}
}

func (n *Node) jobUpdateSessions() error {
	n.Log().Info("Starting job", "name", "update_sessions", "interval", n.IntervalSessions())

	t := time.NewTicker(n.IntervalSessions())
	for ; ; <-t.C {
		var (
			items []*types.Session
		)

		peers, err := n.Service().Peers()
		if err != nil {
			return err
		}
		n.Log().Info("Connected peers", "count", len(peers))

		for _, peer := range peers {
			item := n.Sessions().GetForKey(peer.Key)
			if item == nil {
				n.Log().Error("Unknown connected peer", "info", peer)
				continue
			}

			session, err := n.Client().QuerySession(item.ID)
			if err != nil {
				return err
			}

			var (
				remove, skip = false, false
				consumed     = sdk.NewInt(peer.Upload + peer.Download)
			)

			switch {
			case session.Status.Equal(hubtypes.StatusInactive):
				remove, skip = true, true
				n.Log().Info("Invalid session status", "peer", peer, "item", item, "session", session)
			case peer.Download == session.Bandwidth.Upload.Int64() && session.StatusAt.After(item.ConnectedAt):
				remove, skip = true, true
				n.Log().Info("Stale peer connection", "peer", peer, "item", item, "session", session)
			case consumed.GT(item.Available):
				remove, skip = true, false
				n.Log().Info("Peer quota exceeded", "peer", peer, "item", item, "session", session)
			}

			if remove {
				key, err := base64.StdEncoding.DecodeString(peer.Key)
				if err != nil {
					return err
				}

				if err := n.Service().RemovePeer(key); err != nil {
					return err
				}
				n.Log().Info("Removed peer for underlying service...")

				n.Sessions().DeleteForKey(item.Key)
				n.Sessions().DeleteForAddress(item.Address)
				n.Log().Info("Removed session...")
			}
			if skip {
				continue
			}

			item.Upload = peer.Upload
			item.Download = peer.Download
			items = append(items, item)
		}

		if len(items) == 0 {
			continue
		}
		if err := n.updateSessions(items...); err != nil {
			return err
		}
	}
}
