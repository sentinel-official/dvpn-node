package node

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	hubtypes "github.com/sentinel-official/hub/types"

	"github.com/sentinel-official/dvpn-node/types"
)

func (n *Node) jobSetSessions() error {
	n.Log().Info("Starting job", "name", "set_sessions", "interval", n.IntervalSetSessions())

	t := time.NewTicker(n.IntervalSetSessions())
	for ; ; <-t.C {
		peers, err := n.Service().Peers()
		if err != nil {
			n.Log().Error("Failed to get connected peers", "error", err)
			return err
		}
		n.Log().Info("Connected peers", "count", len(peers))

		for i := 0; i < len(peers); i++ {
			item := n.Sessions().GetByKey(peers[i].Key)
			if item.Empty() {
				n.Log().Error("Unknown connected peer", "peer", peers[i])
				if err := n.RemovePeer(peers[i].Key); err != nil {
					return err
				}

				continue
			}

			item.Upload = peers[i].Upload
			item.Download = peers[i].Download
			n.Sessions().Update(item)

			consumed := sdk.NewInt(item.Upload + item.Download)
			if consumed.GT(item.Available) {
				n.Log().Info("Peer quota exceeded", "id", item.ID,
					"available", item.Available, "consumed", consumed)
				if err := n.RemovePeer(item.Key); err != nil {
					return err
				}
			}
		}
	}
}

func (n *Node) jobUpdateStatus() error {
	n.Log().Info("Starting job", "name", "update_status", "interval", n.IntervalUpdateStatus())

	t := time.NewTicker(n.IntervalUpdateStatus())
	for ; ; <-t.C {
		if err := n.UpdateNodeStatus(); err != nil {
			return err
		}
	}
}

func (n *Node) jobUpdateSessions() error {
	n.Log().Info("Starting job", "name", "update_sessions", "interval", n.IntervalUpdateSessions())

	t := time.NewTicker(n.IntervalUpdateSessions())
	for ; ; <-t.C {
		var items []types.Session
		n.Sessions().Iterate(func(v types.Session) bool {
			items = append(items, v)
			return false
		})
		n.Log().Info("Iterated sessions", "count", len(items))

		for i := len(items) - 1; i >= 0; i-- {
			session, err := n.Client().QuerySession(items[i].ID)
			if err != nil {
				return err
			}

			subscription, err := n.Client().QuerySubscription(session.Subscription)
			if err != nil {
				return err
			}

			remove, skip := func() (bool, bool) {
				switch {
				case !subscription.Status.Equal(hubtypes.StatusActive):
					n.Log().Info("Invalid subscription status", "id", items[i].ID)
					return true, subscription.Status.Equal(hubtypes.StatusInactive)
				case !session.Status.Equal(hubtypes.StatusActive):
					n.Log().Info("Invalid session status", "id", items[i].ID)
					return true, session.Status.Equal(hubtypes.StatusInactive)
				case items[i].Download == session.Bandwidth.Upload.Int64() && items[i].ConnectedAt.Before(session.StatusAt):
					n.Log().Info("Stale peer connection", "id", items[i].ID)
					return true, false
				default:
					return false, false
				}
			}()

			if remove {
				if err := n.RemovePeerAndSession(items[i].Key, items[i].Address); err != nil {
					return err
				}
			}
			if skip {
				items = append(items[:i], items[i+1:]...)
			}
		}

		if len(items) == 0 {
			continue
		}
		if err := n.UpdateSessions(items...); err != nil {
			return err
		}
	}
}
