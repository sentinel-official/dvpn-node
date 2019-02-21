package node

import (
	"encoding/json"
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
	session := types.NewSession(details)
	if n.clients[id] != nil {
		return
	}
	if n.sessions[id] == nil {
		n.sessions[id] = session
		go n.sessionTimeout(id)
	}

	key, err := n.vpn.GenerateClientKey(id)
	if err != nil {
		return
	}

	_, _ = w.Write(key)
}

func (n Node) sessionTimeout(id string) {
	session := n.sessions[id]

	select {
	case <-session.Timeout.C:
		delete(n.clients, id)
		delete(n.sessions, id)
	case <-session.StopTimeout:
		return
	}
}

func (n Node) handleFuncWebsocket(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if n.sessions[id] == nil {
		return
	}
	if n.clients[id] != nil {
		return
	}

	n.sessions[id].StopTimeout <- true
	conn, err := types.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	n.clients[id] = types.NewClient(conn)

	go n.readMessages(id)
	go n.writeMessages(id)
}

func (n Node) readMessages(id string) {
	client := n.clients[id]
	session := n.sessions[id]

	defer func() {
		if err := client.Conn.Close(); err != nil {
			panic(err)
		}

		delete(n.clients, id)
		delete(n.sessions, id)
	}()

	_ = client.Conn.SetReadDeadline(time.Now().Add(30 * time.Second))

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

		_ = client.Conn.SetReadDeadline(time.Now().Add(30 * time.Second))
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

		if err := session.VerifyAndSetBandwidthSigns(data.Upload, data.Download,
			data.NodeOwnerSign, data.ClientSign); err != nil {
			return err
		}
	default:
		return errors.New("Invalid message type")
	}

	return nil
}

func (n Node) writeMessages(id string) {
	client := n.clients[id]

	for {
		msg := <-client.OutMessages
		if err := client.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}
