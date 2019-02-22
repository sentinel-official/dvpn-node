package types

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	Upgrader = &websocket.Upgrader{
		HandshakeTimeout: 45 * time.Second,
	}
)

type Client struct {
	Conn        *websocket.Conn
	OutMessages chan []byte
}

func NewClient(conn *websocket.Conn) *Client {
	return &Client{
		Conn:        conn,
		OutMessages: make(chan []byte),
	}
}

type Clients struct {
	c     map[string]*Client
	mutex *sync.Mutex
}

func NewClients() Clients {
	return Clients{
		c:     make(map[string]*Client),
		mutex: &sync.Mutex{},
	}
}

func (c *Clients) Set(id string, client *Client) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.c[id] = client
}

func (c *Clients) Get(id string) *Client {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.c[id]
}

func (c *Clients) Delete(id string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.c, id)
}
