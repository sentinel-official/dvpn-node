package server

import (
	"encoding/json"
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
	connections map[string]*websocket.Conn
}

func NewServer(db *database.DB, _tx *tx.Tx, upgrader *websocket.Upgrader) *Server {
	return &Server{
		db:          db,
		tx:          _tx,
		upgrader:    upgrader,
		connections: make(map[string]*websocket.Conn),
	}
}

func (s Server) Router() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/websocket", s.HandleWebsocket)

	return router
}

func (s Server) HandleWebsocket(w http.ResponseWriter, r *http.Request) {
	var body webSocketReqBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		panic(err)
	}

	details, err := s.tx.QuerySessionFromTxHash(body.TxHash)
	if err != nil {
		panic(err)
	}

	session := database.NewSessionFromDetails(details)
	if err := s.db.Sessions.AddSession(session); err != nil {
		panic(err)
	}

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		panic(err)
	}

	s.connections[session.SessionID] = conn
}
