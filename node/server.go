package node

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"

	sdkTypes "github.com/ironman0x7b2/sentinel-sdk/types"
	"github.com/ironman0x7b2/sentinel-sdk/x/vpn"

	"github.com/ironman0x7b2/vpn-node/types"
	"github.com/ironman0x7b2/vpn-node/utils"
)

var (
	Upgrader = &websocket.Upgrader{
		HandshakeTimeout: 45 * time.Second,
	}
)

func (n Node) Router() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	router.
		Methods("POST").
		Path("/subscriptions").
		HandlerFunc(n.handlerFuncAddSubscription).
		Name("AddSubscription")

	router.
		Methods("POST").
		Path("/subscriptions/{id}/key").
		HandlerFunc(n.handlerFuncSubscriptionKey).
		Name("SubscriptionKey")

	router.
		Methods("POST").
		Path("/subscriptions/{id}/sessions").
		HandlerFunc(n.handlerFuncInitSession).
		Name("InitSession")

	router.
		Methods("POST").
		Path("/subscriptions/{id}/websocket").
		HandlerFunc(n.handlerFuncSubscriptionWebsocket).
		Name("SubscriptionWebsocket")

	return router
}

type requestAddSubscription struct {
	TxHash string `json:"tx_hash"`
}

func (n Node) handlerFuncAddSubscription(w http.ResponseWriter, r *http.Request) {
	var body requestAddSubscription
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.WriteErrorToResponse(w, 400, types.Error{
			Message: "Error occurred while decoding the response body",
		})
		return
	}

	sub, err := n.tx.QuerySubscriptionByTxHash(body.TxHash)
	if err != nil {
		utils.WriteErrorToResponse(w, 500, types.Error{
			Message: "Error occurred while querying the subscription from chain by transaction hash",
		})
		return
	}
	if sub.Status != vpn.StatusActive {
		utils.WriteErrorToResponse(w, 400, types.Error{
			Message: "Invalid subscription status found on the chain",
		})
		return
	}
	if sub.NodeID != n.id {
		utils.WriteErrorToResponse(w, 400, types.Error{
			Message: "Subscription does not belong to this node",
		})
		return
	}

	query, args := "id = ?", []interface{}{
		sub.ID.String(),
	}

	_sub, err := n.db.SubscriptionFindOne(query, args...)
	if err != nil {
		utils.WriteErrorToResponse(w, 500, types.Error{
			Message: "Error occurred while querying the subscription from database",
		})
		return
	}
	if _sub != nil {
		utils.WriteErrorToResponse(w, 400, types.Error{
			Message: "Subscription is already exists in the database",
		})
		return
	}

	client, err := n.tx.QueryAccount(sub.Client.String())
	if err != nil {
		utils.WriteErrorToResponse(w, 500, types.Error{
			Message: "Error occurred while querying the account from chain",
		})
		return
	}

	_sub = &types.Subscription{
		ID:        sub.ID,
		TxHash:    body.TxHash,
		Address:   client.GetAddress(),
		PubKey:    client.GetPubKey(),
		Bandwidth: sub.RemainingBandwidth,
		Status:    types.ACTIVE,
		CreatedAt: time.Now().UTC(),
	}

	if err := n.db.SubscriptionSave(_sub); err != nil {
		utils.WriteErrorToResponse(w, 500, types.Error{
			Message: "Error occurred while adding the subscription to database",
		})
		return
	}

	utils.WriteResultToResponse(w, 201, nil)
}

func (n Node) handlerFuncSubscriptionKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	query, args := "id = ?", []interface{}{
		vars["id"],
	}

	_sub, err := n.db.SubscriptionFindOne(query, args...)
	if err != nil {
		utils.WriteErrorToResponse(w, 500, types.Error{
			Message: "Error occurred while querying the subscription from database",
		})
		return
	}
	if _sub == nil {
		utils.WriteErrorToResponse(w, 400, types.Error{
			Message: "Subscription does not exist in the database",
		})
		return
	}
	if _sub.Status != types.ACTIVE {
		utils.WriteErrorToResponse(w, 400, types.Error{
			Message: "Invalid subscription status found in the database",
		})
		return
	}
	if !_sub.Bandwidth.AllPositive() {
		utils.WriteErrorToResponse(w, 400, types.Error{
			Message: "Invalid bandwidth found in the database",
		})
		return
	}

	sub, err := n.tx.QuerySubscription(vars["id"])
	if err != nil {
		utils.WriteErrorToResponse(w, 500, types.Error{
			Message: "Error occurred while querying the subscription from chain",
		})
		return
	}
	if sub.Status != vpn.StatusActive {
		utils.WriteErrorToResponse(w, 400, types.Error{
			Message: "Invalid subscription status found on the chain",
		})
		return
	}
	if !sub.RemainingBandwidth.AllPositive() {
		utils.WriteErrorToResponse(w, 400, types.Error{
			Message: "Invalid remaining bandwidth found on the chain",
		})
		return
	}

	// TODO: Revoke previous vpn key

	key, err := n.vpn.GenerateClientKey(vars["id"])
	if err != nil {
		return
	}

	_, _ = w.Write(key)
}

type requestInitSession struct {
	Signature string `json:"signature"`
}

func (n Node) handlerFuncInitSession(w http.ResponseWriter, r *http.Request) {
	var body requestInitSession
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.WriteErrorToResponse(w, 500, types.Error{
			Message: "Error occurred while decoding the body",
		})
		return
	}

	vars := mux.Vars(r)

	query, args := "id = ?", []interface{}{
		vars["id"],
	}

	_sub, err := n.db.SubscriptionFindOne(query, args...)
	if err != nil {
		utils.WriteErrorToResponse(w, 500, types.Error{
			Message: "Error occurred while querying the subscription from database",
		})
		return
	}
	if _sub == nil {
		utils.WriteErrorToResponse(w, 400, types.Error{
			Message: "Subscription does not exist in the database",
		})
		return
	}
	if _sub.Status != types.ACTIVE {
		utils.WriteErrorToResponse(w, 400, types.Error{
			Message: "Invalid subscription status found in the database",
		})
		return
	}
	if !_sub.Bandwidth.AllPositive() {
		utils.WriteErrorToResponse(w, 400, types.Error{
			Message: "Invalid bandwidth found in the database",
		})
		return
	}

	sub, err := n.tx.QuerySubscription(vars["id"])
	if err != nil {
		utils.WriteErrorToResponse(w, 500, types.Error{
			Message: "Error occurred while querying the subscription from chain",
		})
		return
	}
	if sub.Status != vpn.StatusActive {
		utils.WriteErrorToResponse(w, 400, types.Error{
			Message: "Invalid subscription status found on the chain",
		})
		return
	}
	if !sub.RemainingBandwidth.AllPositive() {
		utils.WriteErrorToResponse(w, 400, types.Error{
			Message: "Invalid remaining bandwidth found on the chain",
		})
		return
	}

	index, err := n.tx.QuerySessionsCountOfSubscription(vars["id"])
	if err != nil {
		utils.WriteErrorToResponse(w, 500, types.Error{
			Message: "Error occurred while querying the sessions count of subscription from chain",
		})
		return
	}

	query, args = "id = ? AND index = ?", []interface{}{
		vars["id"],
		index,
	}

	_session, err := n.db.SessionFindOne(query, args...)
	if err != nil {
		utils.WriteErrorToResponse(w, 500, types.Error{
			Message: "Error occurred while querying the session from database",
		})
	}
	if _session == nil {
		_session = &types.Session{
			ID:        sub.ID,
			Index:     index,
			Bandwidth: sdkTypes.NewBandwidthFromInt64(0, 0),
			Signature: nil,
			Status:    types.INACTIVE,
			CreatedAt: time.Now().UTC(),
		}

		if err := n.db.SessionSave(_session); err != nil {
			utils.WriteErrorToResponse(w, 500, types.Error{
				Message: "Error occurred while adding the session to database",
			})
			return
		}
	}
	if _session.Status == types.ACTIVE {
		utils.WriteErrorToResponse(w, 400, types.Error{
			Message: "Session status is active in the database",
		})
		return
	}

	signature, err := base64.StdEncoding.DecodeString(body.Signature)
	if err != nil {
		utils.WriteErrorToResponse(w, 400, types.Error{
			Message: "Error occurred while decoding the signature",
		})
		return
	}

	data := vpn.NewBandwidthSignatureData(_session.ID, _session.Index, _session.Bandwidth)
	if !_sub.PubKey.VerifyBytes(data.Bytes(), signature) {
		utils.WriteErrorToResponse(w, 400, types.Error{
			Message: "Invalid bandwidth signature",
			Info:    _session,
		})
		return
	}

	query, args = "id = ? AND index = ? AND status = ?", []interface{}{
		vars["id"],
		index,
		types.INACTIVE,
	}

	updates := map[string]interface{}{
		"signature": signature,
		"status":    types.INIT,
	}

	if err := n.db.SessionFindOneAndUpdate(updates, query, args...); err != nil {
		utils.WriteErrorToResponse(w, 500, types.Error{
			Message: "Error occurred while updating the session in database",
		})
		return
	}

	utils.WriteResultToResponse(w, 200, _session)
}

func (n Node) handlerFuncSubscriptionWebsocket(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	query, args := "id = ?", []interface{}{
		vars["id"],
	}

	_sub, err := n.db.SubscriptionFindOne(query, args...)
	if err != nil {
		utils.WriteErrorToResponse(w, 500, types.Error{
			Message: "Error occurred while querying the subscription from database",
		})
		return
	}
	if _sub == nil {
		utils.WriteErrorToResponse(w, 400, types.Error{
			Message: "Subscription does not exist in the database",
		})
		return
	}
	if _sub.Status != types.ACTIVE {
		utils.WriteErrorToResponse(w, 400, types.Error{
			Message: "Invalid subscription status found in the database",
		})
		return
	}
	if !_sub.Bandwidth.AllPositive() {
		utils.WriteErrorToResponse(w, 400, types.Error{
			Message: "Invalid bandwidth found in the database",
		})
		return
	}

	sub, err := n.tx.QuerySubscription(vars["id"])
	if err != nil {
		utils.WriteErrorToResponse(w, 500, types.Error{
			Message: "Error occurred while querying the subscription from chain",
		})
		return
	}
	if sub.Status != vpn.StatusActive {
		utils.WriteErrorToResponse(w, 400, types.Error{
			Message: "Invalid subscription status found on the chain",
		})
		return
	}
	if !sub.RemainingBandwidth.AllPositive() {
		utils.WriteErrorToResponse(w, 400, types.Error{
			Message: "Invalid remaining bandwidth found on the chain",
		})
		return
	}

	index, err := n.tx.QuerySessionsCountOfSubscription(vars["id"])
	if err != nil {
		utils.WriteErrorToResponse(w, 500, types.Error{
			Message: "Error occurred while querying the sessions count of subscription from chain",
		})
		return
	}

	query, args = "id = ? AND index = ?", []interface{}{
		vars["id"],
		index,
	}

	_session, err := n.db.SessionFindOne(query, args...)
	if err != nil {
		utils.WriteErrorToResponse(w, 500, types.Error{
			Message: "Error occurred while querying the session from database",
		})
	}
	if _session == nil {
		utils.WriteErrorToResponse(w, 400, types.Error{
			Message: "Session does not exist in the database",
		})
		return
	}
	if _session.Status != types.INIT {
		utils.WriteErrorToResponse(w, 400, types.Error{
			Message: "Invalid session status found in the database",
		})
		return
	}

	query, args = "id = ? AND index = ? AND status = ?", []interface{}{
		vars["id"],
		index,
		types.INIT,
	}

	updates := map[string]interface{}{
		"status": types.ACTIVE,
	}

	if err := n.db.SessionFindOneAndUpdate(updates, query, args...); err != nil {
		utils.WriteErrorToResponse(w, 500, types.Error{
			Message: "Error occurred while updating the session in database",
		})
		return
	}

	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		query, args = "id = ? AND index = ? AND status = ?", []interface{}{
			vars["id"],
			index,
			types.ACTIVE,
		}

		updates := map[string]interface{}{
			"status": types.INIT,
		}

		_ = n.db.SessionFindOneAndUpdate(updates, query, args...)
		return
	}

	n.clients[vars["id"]] = &client{
		pubKey:      _sub.PubKey,
		conn:        conn,
		outMessages: make(chan types.Msg),
	}

	go n.readMessages(vars["id"], index)
	go n.writeMessages(vars["id"], index)
}

func (n Node) readMessages(id string, index uint64) {
	client := n.clients[id]

	defer func() {
		query, args := "id = ? AND index = ? AND status = ?", []interface{}{
			id,
			index,
			types.ACTIVE,
		}

		updates := map[string]interface{}{
			"status": types.INACTIVE,
		}

		if err := n.db.SessionFindOneAndUpdate(updates, query, args...); err != nil {
			panic(err)
		}

		if err := client.conn.Close(); err != nil {
			panic(err)
		}
	}()

	deadline := time.Now().Add(types.ConnectionReadTimeout)
	_ = client.conn.SetReadDeadline(deadline)

	for {
		_, p, err := client.conn.ReadMessage()
		if err != nil {
			return
		}

		var msg types.Msg
		if err := json.Unmarshal(p, &msg); err != nil {
			continue
		}

		if err := n.handleIncomingMessage(client, &msg); err != nil {
			continue
		}

		deadline = time.Now().Add(types.ConnectionReadTimeout)
		_ = client.conn.SetReadDeadline(deadline)
	}
}

func (n Node) handleIncomingMessage(client *client, msg *types.Msg) error {
	switch msg.Type {
	case "MsgBandwidthSignature":
		var data MsgBandwidthSignature
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			return errors.Errorf("Error occurred while decoding the message data")
		}
		if err := data.Validate(); err != nil {
			return err
		}

		_data := vpn.NewBandwidthSignatureData(data.ID, data.Index, data.Bandwidth)
		if n.pubKey.VerifyBytes(_data.Bytes(), data.NodeOwnerSignature) {
			return errors.Errorf("Invalid node owner signature")
		}
		if client.pubKey.VerifyBytes(_data.Bytes(), data.NodeOwnerSignature) {
			return errors.Errorf("Invalid client signature")
		}

		query, args := "id = ? AND index = ? AND status = ? AND upload <= ? AND download <= ?", []interface{}{
			data.ID.String(),
			data.Index,
			types.ACTIVE,
			data.Bandwidth.Upload.Int64(),
			data.Bandwidth.Download.Int64(),
		}

		updates := map[string]interface{}{
			"upload":    data.Bandwidth.Upload.Int64(),
			"download":  data.Bandwidth.Download.Int64(),
			"signature": data.ClientSignature,
		}

		if err := n.db.SessionFindOneAndUpdate(updates, query, args...); err != nil {
			return errors.Errorf("Error occurred while updating the session in database")
		}
	default:
		return errors.Errorf("Invalid message type: %s", msg.Type)
	}

	return nil
}

func (n Node) writeMessages(id string, index uint64) {
	client := n.clients[id]

	for message := range client.outMessages {
		data := message.Bytes()
		if err := client.conn.WriteMessage(websocket.TextMessage, data); err != nil {
			return
		}
	}
}
