package models

import (
	"github.com/gorilla/websocket"
)

type Client struct {
	ID     string
	Socket *websocket.Conn
	Send   chan []byte
	Group  string
}

type Message struct {
	Sender    string `json:"sender"`
	Recipient string `json:"recipient"`
	Content   string `json:"content"`
	Group     string `json:"group"`
}
