package node

import (
	"encoding/base64"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	hubtypes "github.com/sentinel-official/hub/types"

	"github.com/sentinel-official/dvpn-node/types"
)

func (n *Node) jobUpdateStatus() error {
	t := time.NewTicker(n.IntervalStatus())
	for ; ; <-t.C {
		if err := n.updateStatus(); err != nil {
			return err
		}
	}
}

func (n *Node) jobUpdateSessions() error {
	t := time.NewTicker(n.IntervalSessions())
	for ; ; <-t.C {
		var (
			items []*types.Session
		)

		peers, err := n.Service().Peers()
		if err != nil {
			return err
		}

		for _, peer := range peers {
			item := n.Sessions().GetForKey(peer.Key)
			if item == nil {
				continue
			}

			session, err := n.Client().QuerySession(item.ID)
			if err != nil {
				return err
			}

			var (
				consumed = sdk.NewInt(peer.Upload + peer.Download)
				inactive = session.Status.Equal(hubtypes.StatusInactive) ||
					peer.Download == session.Bandwidth.Upload.Int64()
			)

			if inactive || consumed.GT(item.Available) {
				key, err := base64.StdEncoding.DecodeString(peer.Key)
				if err != nil {
					return err
				}
				if err := n.Service().RemovePeer(key); err != nil {
					return err
				}

				n.Sessions().DeleteForKey(item.Key)
				n.Sessions().DeleteForAddress(item.Address)
			}
			if inactive {
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
