package manager

import (
	"encoding/json"
	"log"

	"github.com/Tak1za/ivar/models"
)

type ClientManager struct {
	Broadcast  chan models.Message
	Register   chan *models.Client
	Unregister chan *models.Client
	Groups     map[string]map[*models.Client]bool
}

func NewManager() *ClientManager {
	return &ClientManager{
		Broadcast:  make(chan models.Message, 1),
		Register:   make(chan *models.Client, 1),
		Unregister: make(chan *models.Client, 1),
		Groups:     make(map[string]map[*models.Client]bool),
	}
}

func (manager *ClientManager) Start() {
	for {
		select {
		case conn := <-manager.Register:
			if manager.Groups[conn.Group] == nil {
				manager.Groups[conn.Group] = make(map[*models.Client]bool)
			}
			//Add user to the required group
			manager.Groups[conn.Group][conn] = true
			byteMessage, _ := json.Marshal(&models.Message{Content: "Someone has connected", Group: conn.Group})
			log.Println("Someone has connected")
			manager.Send(byteMessage, conn)
		case conn := <-manager.Unregister:
			currentGroup := manager.Groups[conn.Group]
			if _, ok := currentGroup[conn]; ok {
				//Remove user from the required group
				close(conn.Send)
				delete(currentGroup, conn)
				byteMessage, _ := json.Marshal(&models.Message{Content: "Someone has disconnected", Group: conn.Group})
				log.Println("Someone has disconnected")
				manager.Send(byteMessage, conn)
			}
		case message := <-manager.Broadcast:
			groupId := message.Group
			currentGroup := manager.Groups[groupId]
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
	for conn := range manager.Groups[currentGroup] {
		//Send user connection/disconnection message to all users in the group except the user who connected/disconnected
		if conn != ignore {
			conn.Send <- message
		}
	}
}

func (manager *ClientManager) BroadcastMessage(message models.Message) {
	manager.Broadcast <- message
}

func (manager *ClientManager) UnregisterSubscriber(client *models.Client) {
	manager.Unregister <- client
}

func (manager *ClientManager) RegisterSubscriber(client *models.Client) {
	manager.Register <- client
}
