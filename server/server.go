package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"github.com/ironman0x7b2/vpn-node/database"
	"github.com/ironman0x7b2/vpn-node/tx"
)

type Server struct {
	db          *database.DB
	tx          *tx.Tx
	upgrader    *websocket.Upgrader
	Connections map[string]*Connection
}

func NewServer(db *database.DB, tx *tx.Tx, upgrader *websocket.Upgrader) *Server {
	return &Server{
		db:          db,
		tx:          tx,
		upgrader:    upgrader,
		Connections: make(map[string]*Connection),
	}
}

func (s Server) Router() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/websocket", s.handleWebsocket)

	return router
}

func (s Server) handleWebsocket(w http.ResponseWriter, r *http.Request) {
	var body webSocketReqBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		panic(err)
	}

	details, err := s.tx.QuerySessionFromTxHash(body.TxHash)
	if err != nil {
		panic(err)
	}

	session := database.NewSessionFromDetails(details)
	if err := s.db.Sessions.AddSession(&session); err != nil {
		panic(err)
	}

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		panic(err)
	}

	connection := NewConnection(conn, details)
	s.Connections[session.SessionID] = connection

	go func() {
		if err := s.readMessages(connection); err != nil {
			panic(err)
		}
	}()

	go func() {
		if err := s.writeMessages(connection); err != nil {
			panic(err)
		}
	}()
}

func (s Server) readMessages(connection *Connection) error {
	for {
		_, p, err := connection.Conn.ReadMessage()
		if err != nil {
			return err
		}

		var msg bandwidthSignMsg
		if err := json.Unmarshal(p, &msg); err != nil {
			return err
		}
		if err := msg.Validate(); err != nil {
			return err
		}

		bandwidthSign, err := s.db.Sessions.GetSessionBandwidthSign(msg.ID, msg.Index)
		if err != nil {
			return err
		}
		if bandwidthSign == nil {
			return fmt.Errorf("no bandwidth sign found")
		}
		if msg.Upload != bandwidthSign.Upload || msg.Download != bandwidthSign.Download {
			return fmt.Errorf("upload or download is invalid")
		}

		if err := s.db.Sessions.AddSessionBandwidthClientSign(msg.ID, msg.Index, msg.Sign); err != nil {
			return err
		}
	}
}

func (s Server) writeMessages(connection *Connection) error {
	for {
		msg := <-connection.OutMessages
		if err := connection.Conn.WriteJSON(msg); err != nil {
			return err
		}
	}
}
