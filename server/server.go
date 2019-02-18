package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"github.com/ironman0x7b2/vpn-node/database"
	"github.com/ironman0x7b2/vpn-node/tx"
)

type Server struct {
	DB          *database.DB
	tx          *tx.Tx
	upgrader    *websocket.Upgrader
	Connections map[string]*Connection
}

func NewServer(db *database.DB, tx *tx.Tx, upgrader *websocket.Upgrader) *Server {
	return &Server{
		DB:          db,
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
	txHash := r.URL.Query().Get("txHash")
	details, err := s.tx.QuerySessionFromTxHash(txHash)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	session := database.NewSessionFromDetails(details)
	if err := s.DB.Sessions.AddSession(&session); err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	connection := NewConnection(conn, details)
	s.Connections[session.SessionID] = connection

	go s.readMessages(session.SessionID)
	go s.writeMessages(session.SessionID)
}

func (s Server) readMessages(id string) {
	c := s.Connections[id]

	defer func() {
		if err := c.Conn.Close(); err != nil {
			panic(err)
		}

		close(c.OutMessages)
		delete(s.Connections, id)
	}()

	for {
		_, p, err := c.Conn.ReadMessage()
		if err != nil {
			return
		}

		var msg interface{}
		if err := json.Unmarshal(p, &msg); err != nil {
			c.OutMessages <- NewMsgError(1, err.Error())
			break
		}

		switch msg := msg.(type) {
		case MsgBandwidthSign:
			if err := msg.Validate(); err != nil {
				c.OutMessages <- NewMsgError(2, err.Error())
				break
			}

			bandwidthSign, err := s.DB.Sessions.GetBandwidthSign(msg.SessionID, msg.ID)
			if err != nil {
				c.OutMessages <- NewMsgError(3, err.Error())
				break
			}
			if bandwidthSign == nil {
				c.OutMessages <- NewMsgError(4, err.Error())
				break
			}
			if msg.Upload != bandwidthSign.Upload || msg.Download != bandwidthSign.Download {
				c.OutMessages <- NewMsgError(5, err.Error())
				break
			}

			if err := s.DB.Sessions.AddBandwidthClientSign(msg.SessionID, msg.ID, msg.Sign); err != nil {
				c.OutMessages <- NewMsgError(6, err.Error())
				break
			}
		default:
			c.OutMessages <- NewMsgError(0, "invalid message")
		}
	}
}

func (s Server) writeMessages(id string) {
	c := s.Connections[id]

	for {
		msg := <-c.OutMessages
		if err := c.Conn.WriteJSON(msg); err != nil {
			return
		}
	}
}
