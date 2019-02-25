package node

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	vpnTypes "github.com/ironman0x7b2/sentinel-sdk/x/vpn/types"
	"github.com/pkg/errors"

	"github.com/ironman0x7b2/vpn-node/types"
)

func (n Node) Router() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/keys/{txHash}", n.handleFuncKeys)
	router.HandleFunc("/websocket/{id}", n.handleFuncWebsocket)

	return router
}

func (n Node) handleFuncKeys(w http.ResponseWriter, r *http.Request) {
	txHash := mux.Vars(r)["txHash"]
	details, err := n.tx.QuerySessionDetailsFromTxHash(txHash)
	if err != nil {
		return
	}
	if details.NodeID.String() != n.ID.String() {
		return
	}
	if details.Status != vpnTypes.StatusInit {
		return
	}

	id := details.ID.String()
	if n.clients.Get(id) != nil {
		return
	}
	if n.sessions.Get(id) == nil {
		n.sessions.Set(id, types.NewSession(details))
		go n.startSessionTimeout(id)
	}

	key, err := n.vpn.GenerateClientKey(id)
	if err != nil {
		return
	}

	_, _ = w.Write(key)
}

func (n Node) startSessionTimeout(id string) {
	session := n.sessions.Get(id)

	select {
	case <-time.After(types.TimeoutSession):
		n.clients.Delete(id)
		n.sessions.Delete(id)
	case <-session.StopTimeout:
		return
	}
}

func (n Node) handleFuncWebsocket(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	session := n.sessions.Get(id)
	if session == nil {
		return
	}
	if n.clients.Get(id) != nil {
		return
	}

	session.StopTimeout <- true

	conn, err := types.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	n.clients.Set(id, types.NewClient(conn))

	go n.readMessages(id)
	go n.writeMessages(id)
}

func (n Node) readMessages(id string) {
	client := n.clients.Get(id)
	session := n.sessions.Get(id)

	defer func() {
		if err := client.Conn.Close(); err != nil {
			panic(err)
		}

		n.clients.Delete(id)
		n.sessions.Delete(id)
	}()

	_ = client.Conn.SetReadDeadline(time.Now().Add(types.TimeoutConnectionRead))

	for {
		_, p, err := client.Conn.ReadMessage()
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

		_ = client.Conn.SetReadDeadline(time.Now().Add(types.TimeoutConnectionRead))
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

		if err := session.VerifyAndSetBandwidthInfo(data.Bandwidth, data.NodeOwnerSign,
			data.ClientSign); err != nil {
			return err
		}
	default:
		return errors.New(fmt.Sprintf("Invalid message type: %s", msg.Type))
	}

	return nil
}

func (n Node) writeMessages(id string) {
	client := n.clients.Get(id)

	for {
		msg := <-client.OutMessages
		if err := client.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}
