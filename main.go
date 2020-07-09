package main

import (
	"encoding/json"
	"log"
	"os"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
)

type ClientManager struct {
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client
	groups     map[string]map[*Client]bool
}

type Client struct {
	id     string
	socket *websocket.Conn
	send   chan []byte
	group  string
}

type Message struct {
	Sender    string `json:"sender"`
	Recipient string `json:"recipient"`
	Content   string `json:"content"`
	Group     string `json:"group"`
}

var manager = ClientManager{
	broadcast:  make(chan Message),
	register:   make(chan *Client),
	unregister: make(chan *Client),
	groups:     make(map[string]map[*Client]bool),
}

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (manager *ClientManager) start() {
	for {
		select {
		case conn := <-manager.register:
			if manager.groups[conn.group] == nil {
				manager.groups[conn.group] = make(map[*Client]bool)
			}
			//Add user to the required group
			manager.groups[conn.group][conn] = true
			byteMessage, _ := json.Marshal(&Message{Content: "Someone has connected", Group: conn.group})
			log.Println("Someone has connected")
			manager.send(byteMessage, conn)
		case conn := <-manager.unregister:
			currentGroup := manager.groups[conn.group]
			if _, ok := currentGroup[conn]; ok {
				//Remove user from the required group
				close(conn.send)
				delete(currentGroup, conn)
				byteMessage, _ := json.Marshal(&Message{Content: "Someone has disconnected", Group: conn.group})
				log.Println("Someone has disconnected")
				manager.send(byteMessage, conn)
			}
		case message := <-manager.broadcast:
			groupId := message.Group
			currentGroup := manager.groups[groupId]
			byteMessage, _ := json.Marshal(&message)
			//Send message to only users of that group
			for conn := range currentGroup {
				select {
				case conn.send <- byteMessage:
				default:
					close(conn.send)
					delete(currentGroup, conn)
				}
			}
		}
	}
}

func (manager *ClientManager) send(message []byte, ignore *Client) {
	currentGroup := ignore.group
	for conn := range manager.groups[currentGroup] {
		//Send user connection/disconnection message to all users in the group except the user who connected/disconnected
		if conn != ignore {
			conn.send <- message
		}
	}
}

func (c *Client) read() {
	defer func() {
		manager.unregister <- c
		c.socket.Close()
	}()

	for {
		_, message, err := c.socket.ReadMessage()
		if err != nil {
			manager.unregister <- c
			c.socket.Close()
			break
		}
		jsonMessage := Message{Sender: c.id, Content: string(message), Group: c.group}
		manager.broadcast <- jsonMessage
	}
}

func (c *Client) write() {
	defer func() {
		c.socket.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.socket.WriteMessage(websocket.TextMessage, message)
		}
	}
}

func checkOrigin(r *http.Request) func(r *http.Request) bool {
	return func(r *http.Request) bool {
		return true
	}
}

func wsHandler(c *gin.Context) {
	wsUpgrader.CheckOrigin = checkOrigin(c.Request)
	wsUpgrader.EnableCompression = true
	groupId := c.Param("id")
	conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade to websocket: %+v", err)
		return
	}

	client := &Client{id: uuid.NewV4().String(), socket: conn, send: make(chan []byte), group: groupId}

	manager.register <- client

	go client.read()
	go client.write()
}

func main() {
	r := gin.Default()
	go manager.start()

	r.GET("/ws/:id", wsHandler)
	r.Run(":" + os.Getenv("PORT"))
}
