package subscriber

import (
	imanager "github.com/Tak1za/ivar/interface"
	"github.com/Tak1za/ivar/models"
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
)

type MyClient struct {
	Client *models.Client
}

func NewSubscriber(socketConn *websocket.Conn, groupId string) *MyClient {
	myClient := &models.Client{
		ID:     uuid.NewV4().String(),
		Socket: socketConn,
		Send:   make(chan []byte, 1),
		Group:  groupId,
	}

	return &MyClient{Client: myClient}
}

func (c *MyClient) Read(mgr imanager.IManager) {
	defer func() {
		mgr.UnregisterSubscriber(c.Client)
		c.Client.Socket.Close()
	}()

	for {
		_, message, err := c.Client.Socket.ReadMessage()
		if err != nil {
			mgr.UnregisterSubscriber(c.Client)
			c.Client.Socket.Close()
			break
		}
		jsonMessage := models.Message{Sender: c.Client.ID, Content: string(message), Group: c.Client.Group}
		mgr.BroadcastMessage(jsonMessage)
	}
}

func (c *MyClient) Write(mgr imanager.IManager) {
	defer func() {
		mgr.UnregisterSubscriber(c.Client)
		c.Client.Socket.Close()
	}()

	for {
		select {
		case message, ok := <-c.Client.Send:
			if !ok {
				mgr.UnregisterSubscriber(c.Client)
				c.Client.Socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.Client.Socket.WriteMessage(websocket.TextMessage, message)
		}
	}
}
