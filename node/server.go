package node

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/ironman0x7b2/sentinel-sdk/x/vpn"
	"github.com/pkg/errors"

	"github.com/ironman0x7b2/vpn-node/types"
)

func (n Node) Router() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/keys/{hash}", n.handleFuncKeys)
	router.HandleFunc("/websocket/{id}", n.handleFuncWebsocket)

	return router
}

func (n Node) handleFuncKeys(w http.ResponseWriter, r *http.Request) {
	hash := mux.Vars(r)["hash"]

	_session, err := n.tx.QuerySessionFromTxHash(hash)
	if err != nil {
		return
	}
	if _session.NodeID.String() != n.id.String() {
		return
	}
	if !_session.NodeOwner.Equals(n.owner) {
		return
	}
	if _session.Status != vpn.StatusInit {
		return
	}

	id := _session.ID.HashTruncated()
	session := n.sessions.Get(id)

	if session == nil {
		session = types.NewSession(_session)
		n.sessions.Set(id, session)

		go n.listenTimeout(id)
	}

	if session.Status != vpn.StatusInit {
		return
	}

	key, err := n.vpn.GenerateClientKey(id)
	if err != nil {
		return
	}

	_, _ = w.Write(key)
}

func (n Node) listenTimeout(id string) {
	session := n.sessions.Get(id)

	select {
	case <-session.Timeout():
		session.Status = vpn.StatusInactive

		if err := n.vpn.DisconnectClient(id); err != nil {
			panic(err)
		}
	case <-session.StopTimeoutListener():
		return
	}
}

func (n Node) handleFuncWebsocket(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	session := n.sessions.Get(id)
	if session == nil {
		return
	}
	if session.Status != vpn.StatusInit {
		return
	}

	session.StopTimeout <- true

	conn, err := types.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	session.Conn = conn
	session.Status = vpn.StatusActive

	go n.readMessages(id)
	go n.writeMessages(id)
}

func (n Node) readMessages(id string) {
	session := n.sessions.Get(id)

	defer func() {
		session.Status = vpn.StatusInactive

		if err := session.Conn.Close(); err != nil {
			panic(err)
		}

		if err := n.vpn.DisconnectClient(id); err != nil {
			panic(err)
		}
	}()

	_ = session.Conn.SetReadDeadline(time.Now().Add(types.ConnectionReadTimeout))

	for {
		_, p, err := session.Conn.ReadMessage()
		if err != nil {
			return
		}

		var msg types.Msg
		if err := json.Unmarshal(p, &msg); err != nil {
			continue
		}

		if err := n.handleIncomingMessage(session, &msg); err != nil {
			continue
		}

		_ = session.Conn.SetReadDeadline(time.Now().Add(types.ConnectionReadTimeout))
	}
}

func (n Node) handleIncomingMessage(session *types.Session, msg *types.Msg) error {
	switch msg.Type {
	case "msg_bandwidth_sign":
		var data MsgBandwidthSign
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			return err
		}
		if err := data.Validate(); err != nil {
			return err
		}

		if err := session.VerifyAndSetConsumedBandwidth(data.Bandwidth,
			data.NodeOwnerSign, data.ClientSign); err != nil {

			return err
		}
	default:
		return errors.Errorf("Invalid message type: %s", msg.Type)
	}

	return nil
}

func (n Node) writeMessages(id string) {
	session := n.sessions.Get(id)
	for message := range session.OutMessages {
		data := message.GetBytes()

		err := session.Conn.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			return
		}
	}
}
