package ws

import "github.com/9thDuck/chat_go.git/internal/store"

type MessageEvent struct {
	Message store.Message `json:"message"`
	Type    string        `json:"type"`
}
const (
	EVENT_MESSAGE                = "MESSAGE"
)
