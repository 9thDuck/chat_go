package ws

import (
	"fmt"
	"sync"
)

type Hub struct {
	clients          map[*Client]bool
	clientsWithIDKey map[int64]*Client
	register         chan *Client
	unregister       chan *Client
	broadcast        chan []byte
	writeToClient    chan []byte
	sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:          make(map[*Client]bool),
		clientsWithIDKey: make(map[int64]*Client),
		register:         make(chan *Client),
		unregister:       make(chan *Client),
		broadcast:        make(chan []byte),
		writeToClient:    make(chan []byte),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.Lock()
			h.clients[client] = true
			h.clientsWithIDKey[client.id] = client
			client.send <- []byte("Welcome to the chat!")
			h.Unlock()
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				h.Lock()
				delete(h.clients, client)
				delete(h.clientsWithIDKey, client.id)
				fmt.Println("client unregistered", client.id)
				client.send <- []byte("You have been disconnected from the chat.")
				client.conn.Close()
				h.Unlock()
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				client.send <- message
			}
		}
	}
}

func (h *Hub) WriteToClient(receiverID int64, message []byte) bool {
	client, ok := h.clientsWithIDKey[receiverID]
	if !ok {
		fmt.Printf("writing to socket client failed, client not found: %v", receiverID)
		return false
	}
	client.send <- message
	return true
}
