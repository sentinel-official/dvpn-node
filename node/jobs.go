package node

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	hubtypes "github.com/sentinel-official/hub/types"

	"github.com/sentinel-official/dvpn-node/types"
)

func (n *Node) jobSetSessions() error {
	n.Log().Info("Starting a job", "name", "set_sessions", "interval", n.IntervalSetSessions())

	t := time.NewTicker(n.IntervalSetSessions())
	for ; ; <-t.C {
		peers, err := n.Service().Peers()
		if err != nil {
			return err
		}

		for i := 0; i < len(peers); i++ {
			var item types.Session
			n.Database().Where(
				&types.Session{
					Key: peers[i].Key,
				},
			).First(&item)

			if item.ID == 0 {
				n.Log().Info("Unknown connected peer", "key", peers[i].Key)
				if err := n.RemovePeer(peers[i].Key); err != nil {
					return err
				}

				continue
			}

			n.Database().Model(
				&types.Session{
					Key: peers[i].Key,
				},
			).Updates(
				&types.Session{
					Upload:   peers[i].Upload,
					Download: peers[i].Download,
				},
			)

			var (
				available = sdk.NewInt(item.Available)
				consumed  = sdk.NewInt(peers[i].Upload + peers[i].Download)
			)

			if consumed.GT(available) {
				n.Log().Info("Peer quota exceeded", "key", peers[i].Key)
				if err := n.RemovePeer(item.Key); err != nil {
					return err
				}
			}
		}
	}
}

func (n *Node) jobUpdateStatus() error {
	n.Log().Info("Starting a job", "name", "update_status", "interval", n.IntervalUpdateStatus())

	t := time.NewTicker(n.IntervalUpdateStatus())
	for ; ; <-t.C {
		if err := n.UpdateNodeStatus(); err != nil {
			return err
		}
	}
}

func (n *Node) jobUpdateSessions() error {
	n.Log().Info("Starting a job", "name", "update_sessions", "interval", n.IntervalUpdateSessions())

	t := time.NewTicker(n.IntervalUpdateSessions())
	for ; ; <-t.C {
		var items []types.Session
		n.Database().Find(&items)

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
				var (
					nochange = items[i].Download == session.Bandwidth.Upload.Int64()
				)

				switch {
				case nochange && items[i].CreatedAt.Before(session.StatusAt): // TODO: Review condition here
					n.Log().Info("Stale peer connection", "id", items[i].ID)
					return true, true
				case !subscription.Status.Equal(hubtypes.StatusActive):
					n.Log().Info("Invalid subscription status", "id", items[i].ID, "nochange", nochange)
					return true, nochange || subscription.Status.Equal(hubtypes.StatusInactive)
				case !session.Status.Equal(hubtypes.StatusActive):
					n.Log().Info("Invalid session status", "id", items[i].ID, "nochange", nochange)
					return true, nochange || session.Status.Equal(hubtypes.StatusInactive)
				default:
					return false, false
				}
			}()

			if remove {
				if err := n.RemovePeer(items[i].Key); err != nil {
					return err
				}

				n.Database().Delete(
					&types.Session{
						Address: items[i].Address,
					},
				)
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
