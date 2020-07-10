package manager

import (
	"encoding/json"
	"log"

	"github.com/Tak1za/ivar/models"
)

type ClientManager struct {
	broadcast  chan models.Message
	register   chan *models.Client
	unregister chan *models.Client
	groups     map[string]map[*models.Client]bool
}

func NewManager() *ClientManager {
	return &ClientManager{
		broadcast:  make(chan models.Message),
		register:   make(chan *models.Client),
		unregister: make(chan *models.Client),
		groups:     make(map[string]map[*models.Client]bool),
	}
}

func (manager *ClientManager) Start() {
	for {
		select {
		case conn := <-manager.register:
			if manager.groups[conn.Group] == nil {
				manager.groups[conn.Group] = make(map[*models.Client]bool)
			}
			//Add user to the required group
			manager.groups[conn.Group][conn] = true
			byteMessage, _ := json.Marshal(&models.Message{Content: "Someone has connected", Group: conn.Group})
			log.Println("Someone has connected")
			manager.Send(byteMessage, conn)
		case conn := <-manager.unregister:
			currentGroup := manager.groups[conn.Group]
			if _, ok := currentGroup[conn]; ok {
				//Remove user from the required group
				close(conn.Send)
				delete(currentGroup, conn)
				byteMessage, _ := json.Marshal(&models.Message{Content: "Someone has disconnected", Group: conn.Group})
				log.Println("Someone has disconnected")
				manager.Send(byteMessage, conn)
			}
		case message := <-manager.broadcast:
			groupId := message.Group
			currentGroup := manager.groups[groupId]
			byteMessage, _ := json.Marshal(&message)
			//Send message to only users of that group
			for conn := range currentGroup {
				select {
				case conn.Send <- byteMessage:
				default:
					close(conn.Send)
					delete(currentGroup, conn)
				}
			}
		}
	}
}

func (manager *ClientManager) Send(message []byte, ignore *models.Client) {
	currentGroup := ignore.Group
	for conn := range manager.groups[currentGroup] {
		//Send user connection/disconnection message to all users in the group except the user who connected/disconnected
		if conn != ignore {
			conn.Send <- message
		}
	}
}

func (manager *ClientManager) BroadcastMessage(message models.Message) {
	manager.broadcast <- message
}

func (manager *ClientManager) UnregisterSubscriber(client *models.Client) {
	manager.unregister <- client
}

func (manager *ClientManager) RegisterSubscriber(client *models.Client) {
	manager.register <- client
}
