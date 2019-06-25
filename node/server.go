// nolint:gocyclo,gochecknoglobals
package node

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/tendermint/tendermint/crypto"

	hub "github.com/sentinel-official/sentinel-hub/types"
	"github.com/sentinel-official/sentinel-hub/x/vpn"

	"github.com/sentinel-official/sentinel-dvpn-node/types"
	"github.com/sentinel-official/sentinel-dvpn-node/utils"
)

type client struct {
	pubKey      crypto.PubKey
	conn        *websocket.Conn
	outMessages chan *types.Msg
}

var (
	upgrader = &websocket.Upgrader{
		HandshakeTimeout: 45 * time.Second,
	}
)

func (n *Node) Router() *mux.Router {
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

func (n *Node) handlerFuncAddSubscription(w http.ResponseWriter, r *http.Request) {
	var body struct {
		TxHash string `json:"tx_hash"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.WriteErrorToResponse(w, 400, &types.StdError{
			Message: "Error occurred while decoding the response body",
			Info:    err.Error(),
		})
		return
	}

	sub, err := n.tx.QuerySubscriptionByTxHash(body.TxHash)
	if err != nil {
		utils.WriteErrorToResponse(w, 500, &types.StdError{
			Message: "Error occurred while querying the subscription from chain by transaction hash",
			Info:    err.Error(),
		})
		return
	}
	if sub.Status != vpn.StatusActive {
		utils.WriteErrorToResponse(w, 400, &types.StdError{
			Message: "Invalid subscription status found on the chain",
			Info:    sub,
		})
		return
	}
	if sub.NodeID != n.id {
		utils.WriteErrorToResponse(w, 400, &types.StdError{
			Message: "Subscription does not belong to this node",
			Info:    sub,
		})
		return
	}

	query, args := "_id = ?", []interface{}{
		sub.ID.String(),
	}

	_sub, err := n.db.SubscriptionFindOne(query, args...)
	if err != nil {
		utils.WriteErrorToResponse(w, 500, &types.StdError{
			Message: "Error occurred while querying the subscription from database",
			Info:    err.Error(),
		})
		return
	}
	if _sub != nil {
		utils.WriteErrorToResponse(w, 400, &types.StdError{
			Message: "Subscription is already exists in the database",
			Info:    _sub,
		})
		return
	}

	client, err := n.tx.QueryAccount(sub.Client.String())
	if err != nil {
		utils.WriteErrorToResponse(w, 500, &types.StdError{
			Message: "Error occurred while querying the account from chain",
			Info:    err.Error(),
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
		utils.WriteErrorToResponse(w, 500, &types.StdError{
			Message: "Error occurred while adding the subscription to database",
			Info:    err.Error(),
		})
		return
	}

	utils.WriteResultToResponse(w, 201, _sub)
}

func (n *Node) handlerFuncSubscriptionKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	query, args := "_id = ?", []interface{}{
		vars["id"],
	}

	_sub, err := n.db.SubscriptionFindOne(query, args...)
	if err != nil {
		utils.WriteErrorToResponse(w, 500, &types.StdError{
			Message: "Error occurred while querying the subscription from database",
			Info:    err.Error(),
		})
		return
	}
	if _sub == nil {
		utils.WriteErrorToResponse(w, 400, &types.StdError{
			Message: "Subscription does not exist in the database",
		})
		return
	}
	if _sub.Status != types.ACTIVE {
		utils.WriteErrorToResponse(w, 400, &types.StdError{
			Message: "Invalid subscription status found in the database",
			Info:    _sub,
		})
		return
	}
	if !_sub.Bandwidth.AllPositive() {
		utils.WriteErrorToResponse(w, 400, &types.StdError{
			Message: "Invalid bandwidth found in the database",
			Info:    _sub,
		})
		return
	}

	sub, err := n.tx.QuerySubscription(vars["id"])
	if err != nil {
		utils.WriteErrorToResponse(w, 500, &types.StdError{
			Message: "Error occurred while querying the subscription from chain",
			Info:    err.Error(),
		})
		return
	}
	if sub.Status != vpn.StatusActive {
		utils.WriteErrorToResponse(w, 400, &types.StdError{
			Message: "Invalid subscription status found on the chain",
			Info:    sub,
		})
		return
	}
	if !sub.RemainingBandwidth.AllPositive() {
		utils.WriteErrorToResponse(w, 400, &types.StdError{
			Message: "Invalid remaining bandwidth found on the chain",
			Info:    sub,
		})
		return
	}

	_ = n.vpn.RevokeClient(vars["id"])

	key, err := n.vpn.GenerateClientKey(vars["id"])
	if err != nil {
		utils.WriteErrorToResponse(w, 400, &types.StdError{
			Message: "Error occurred while generating client key",
			Info:    err.Error(),
		})
		return
	}

	utils.WriteResultToResponse(w, 200, key)
}

func (n *Node) handlerFuncInitSession(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Signature string `json:"signature"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.WriteErrorToResponse(w, 500, &types.StdError{
			Message: "Error occurred while decoding the body",
			Info:    err.Error(),
		})
		return
	}

	signature, err := base64.StdEncoding.DecodeString(body.Signature)
	if err != nil {
		utils.WriteErrorToResponse(w, 400, &types.StdError{
			Message: "Error occurred while decoding the signature",
			Info:    err.Error(),
		})
		return
	}

	vars := mux.Vars(r)

	query, args := "_id = ?", []interface{}{
		vars["id"],
	}

	_sub, err := n.db.SubscriptionFindOne(query, args...)
	if err != nil {
		utils.WriteErrorToResponse(w, 500, &types.StdError{
			Message: "Error occurred while querying the subscription from database",
			Info:    err.Error(),
		})
		return
	}
	if _sub == nil {
		utils.WriteErrorToResponse(w, 400, &types.StdError{
			Message: "Subscription does not exist in the database",
		})
		return
	}
	if _sub.Status != types.ACTIVE {
		utils.WriteErrorToResponse(w, 400, &types.StdError{
			Message: "Invalid subscription status found in the database",
			Info:    _sub,
		})
		return
	}
	if !_sub.Bandwidth.AllPositive() {
		utils.WriteErrorToResponse(w, 400, &types.StdError{
			Message: "Invalid bandwidth found in the database",
			Info:    _sub,
		})
		return
	}

	sub, err := n.tx.QuerySubscription(vars["id"])
	if err != nil {
		utils.WriteErrorToResponse(w, 500, &types.StdError{
			Message: "Error occurred while querying the subscription from chain",
			Info:    err.Error(),
		})
		return
	}
	if sub.Status != vpn.StatusActive {
		utils.WriteErrorToResponse(w, 400, &types.StdError{
			Message: "Invalid subscription status found on the chain",
			Info:    sub,
		})
		return
	}
	if !sub.RemainingBandwidth.AllPositive() {
		utils.WriteErrorToResponse(w, 400, &types.StdError{
			Message: "Invalid remaining bandwidth found on the chain",
			Info:    sub,
		})
		return
	}

	index, err := n.tx.QuerySessionsCountOfSubscription(vars["id"])
	if err != nil {
		utils.WriteErrorToResponse(w, 500, &types.StdError{
			Message: "Error occurred while querying the sessions count of subscription from chain",
			Info:    err.Error(),
		})
		return
	}

	query, args = "_id = ? AND _index = ?", []interface{}{
		vars["id"],
		index,
	}

	_session, err := n.db.SessionFindOne(query, args...)
	if err != nil {
		utils.WriteErrorToResponse(w, 500, &types.StdError{
			Message: "Error occurred while querying the session from database",
			Info:    err.Error(),
		})
	}
	if _session == nil {
		_session = &types.Session{
			ID:        sub.ID,
			Index:     index,
			Bandwidth: hub.NewBandwidthFromInt64(0, 0),
			Signature: nil,
			Status:    types.INIT,
			CreatedAt: time.Now().UTC(),
		}

		if err = n.db.SessionSave(_session); err != nil {
			utils.WriteErrorToResponse(w, 500, &types.StdError{
				Message: "Error occurred while adding the session to database",
				Info:    err.Error(),
			})
			return
		}

		utils.WriteResultToResponse(w, 201, _session)
		return
	}
	if _session.Status == types.ACTIVE {
		utils.WriteErrorToResponse(w, 400, &types.StdError{
			Message: "Session status is active in the database",
			Info:    _session,
		})
		return
	}

	data := vpn.NewBandwidthSignatureData(_session.ID, _session.Index, _session.Bandwidth)
	if !_sub.PubKey.VerifyBytes(data.Bytes(), signature) {
		utils.WriteErrorToResponse(w, 400, &types.StdError{
			Message: "Invalid bandwidth signature",
			Info:    _session,
		})
		return
	}

	query, args = "_id = ? AND _index = ? AND _status IN (?)", []interface{}{
		vars["id"],
		index,
		[]string{types.INIT, types.INACTIVE},
	}

	updates := map[string]interface{}{
		"_signature": signature,
		"_status":    types.INIT,
	}

	if err := n.db.SessionFindOneAndUpdate(updates, query, args...); err != nil {
		utils.WriteErrorToResponse(w, 500, &types.StdError{
			Message: "Error occurred while updating the session in database",
			Info:    err.Error(),
		})
		return
	}

	utils.WriteResultToResponse(w, 200, _session)
}

func (n *Node) handlerFuncSubscriptionWebsocket(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	query, args := "_id = ?", []interface{}{
		vars["id"],
	}

	_sub, err := n.db.SubscriptionFindOne(query, args...)
	if err != nil {
		utils.WriteErrorToResponse(w, 500, &types.StdError{
			Message: "Error occurred while querying the subscription from database",
			Info:    err.Error(),
		})
		return
	}
	if _sub == nil {
		utils.WriteErrorToResponse(w, 400, &types.StdError{
			Message: "Subscription does not exist in the database",
		})
		return
	}
	if _sub.Status != types.ACTIVE {
		utils.WriteErrorToResponse(w, 400, &types.StdError{
			Message: "Invalid subscription status found in the database",
			Info:    _sub,
		})
		return
	}
	if !_sub.Bandwidth.AllPositive() {
		utils.WriteErrorToResponse(w, 400, &types.StdError{
			Message: "Invalid bandwidth found in the database",
			Info:    _sub,
		})
		return
	}

	sub, err := n.tx.QuerySubscription(vars["id"])
	if err != nil {
		utils.WriteErrorToResponse(w, 500, &types.StdError{
			Message: "Error occurred while querying the subscription from chain",
			Info:    err.Error(),
		})
		return
	}
	if sub.Status != vpn.StatusActive {
		utils.WriteErrorToResponse(w, 400, &types.StdError{
			Message: "Invalid subscription status found on the chain",
			Info:    sub,
		})
		return
	}
	if !sub.RemainingBandwidth.AllPositive() {
		utils.WriteErrorToResponse(w, 400, &types.StdError{
			Message: "Invalid remaining bandwidth found on the chain",
			Info:    sub,
		})
		return
	}

	index, err := n.tx.QuerySessionsCountOfSubscription(vars["id"])
	if err != nil {
		utils.WriteErrorToResponse(w, 500, &types.StdError{
			Message: "Error occurred while querying the sessions count of subscription from chain",
			Info:    err.Error(),
		})
		return
	}

	query, args = "_id = ? AND _index = ?", []interface{}{
		vars["id"],
		index,
	}

	_session, err := n.db.SessionFindOne(query, args...)
	if err != nil {
		utils.WriteErrorToResponse(w, 500, &types.StdError{
			Message: "Error occurred while querying the session from database",
			Info:    err.Error(),
		})
	}
	if _session == nil {
		utils.WriteErrorToResponse(w, 400, &types.StdError{
			Message: "Session does not exist in the database",
		})
		return
	}
	if _session.Status != types.INIT {
		utils.WriteErrorToResponse(w, 400, &types.StdError{
			Message: "Invalid session status found in the database",
			Info:    _session,
		})
		return
	}

	query, args = "_id = ? AND _index = ? AND _status = ?", []interface{}{
		vars["id"],
		index,
		types.INIT,
	}

	updates := map[string]interface{}{
		"_status": types.ACTIVE,
	}

	if err = n.db.SessionFindOneAndUpdate(updates, query, args...); err != nil {
		utils.WriteErrorToResponse(w, 500, &types.StdError{
			Message: "Error occurred while updating the session in database",
			Info:    err.Error(),
		})
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		query, args = "_id = ? AND _index = ? AND _status = ?", []interface{}{
			vars["id"],
			index,
			types.ACTIVE,
		}

		updates = map[string]interface{}{
			"_status": types.INIT,
		}

		_ = n.db.SessionFindOneAndUpdate(updates, query, args...)
		return
	}

	n.clients[vars["id"]] = &client{
		pubKey:      _sub.PubKey,
		conn:        conn,
		outMessages: make(chan *types.Msg),
	}

	go n.readMessages(vars["id"], index)
	go n.writeMessages(vars["id"])
}

func (n *Node) readMessages(id string, index uint64) {
	client := n.clients[id]

	defer func() {
		query, args := "_id = ? AND _index = ? AND _status = ?", []interface{}{
			id,
			index,
			types.ACTIVE,
		}

		updates := map[string]interface{}{
			"_status": types.INACTIVE,
		}

		if err := n.db.SessionFindOneAndUpdate(updates, query, args...); err != nil {
			panic(err)
		}

		if err := client.conn.Close(); err != nil {
			panic(err)
		}
	}()

	_ = client.conn.SetReadDeadline(
		time.Now().Add(types.ConnectionReadTimeout))

	for {
		_, p, err := client.conn.ReadMessage()
		if err != nil {
			return
		}

		var msg types.Msg
		if err := json.Unmarshal(p, &msg); err != nil {
			client.outMessages <- NewMsgError(1, "Error occurred while decoding the message")
			continue
		}

		if errMsg := n.handleIncomingMessage(client.pubKey, &msg); errMsg != nil {
			client.outMessages <- errMsg
			continue
		}

		_ = client.conn.SetReadDeadline(
			time.Now().Add(types.ConnectionReadTimeout))
	}
}

func (n *Node) handleIncomingMessage(pubKey crypto.PubKey, msg *types.Msg) *types.Msg {
	switch msg.Type {
	case "MsgBandwidthSignature":
		return n.handleMsgBandwidthSignature(pubKey, msg.Data)
	default:
		return NewMsgError(1, "Invalid message type")
	}
}

func (n *Node) handleMsgBandwidthSignature(pubKey crypto.PubKey, rawMsg json.RawMessage) *types.Msg {
	var msg MsgBandwidthSignature
	if err := json.Unmarshal(rawMsg, &msg); err != nil {
		return NewMsgError(2, "Error occurred while decoding the raw message")
	}
	if err := msg.Validate(); err != nil {
		return NewMsgError(3, "Invalid message")
	}

	data := vpn.NewBandwidthSignatureData(msg.ID, msg.Index, msg.Bandwidth).Bytes()
	if !n.pubKey.VerifyBytes(data, msg.NodeOwnerSignature) {
		return NewMsgError(4, "Invalid node owner signature")
	}
	if !pubKey.VerifyBytes(data, msg.ClientSignature) {
		return NewMsgError(5, "Invalid client signature")
	}

	query, args := "_id = ? AND _index = ? AND _status = ? AND _upload <= ? AND _download <= ?", []interface{}{
		msg.ID.String(),
		msg.Index,
		types.ACTIVE,
		msg.Bandwidth.Upload.Int64(),
		msg.Bandwidth.Download.Int64(),
	}

	updates := map[string]interface{}{
		"_upload":    msg.Bandwidth.Upload.Int64(),
		"_download":  msg.Bandwidth.Download.Int64(),
		"_signature": msg.ClientSignature,
	}

	if err := n.db.SessionFindOneAndUpdate(updates, query, args...); err != nil {
		return NewMsgError(6, "Error occurred while updating the session in database")
	}

	return nil
}

func (n *Node) writeMessages(id string) {
	client := n.clients[id]

	for message := range client.outMessages {
		data := message.Bytes()
		if err := client.conn.WriteMessage(websocket.TextMessage, data); err != nil {
			return
		}
	}
}
