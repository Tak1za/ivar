package manager_test

import (
	"github.com/Tak1za/ivar/manager"
	"github.com/Tak1za/ivar/models"
	"github.com/Tak1za/ivar/subscriber"
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
	"net/http/httptest"
	"testing"
)

var mockManager *manager.ClientManager
var mockClient *subscriber.MyClient

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func init() {
	mockManager = manager.NewManager()
	r := httptest.NewRequest("GET", "/foo", nil)
	w := httptest.NewRecorder()
	mockSocketConn, _ := wsUpgrader.Upgrade(w, r, nil)
	mockClient = subscriber.NewSubscriber(mockSocketConn, "123")
}

func TestClientManager_BroadcastMessage(t *testing.T) {
	mockMessage := models.Message{
		Sender:    uuid.NewV4().String(),
		Recipient: uuid.NewV4().String(),
		Content:   "Hi, how are you?",
		Group:     "123",
	}

	mockManager.BroadcastMessage(mockMessage)

	gotMessage := <-mockManager.Broadcast
	if gotMessage != mockMessage {
		t.Errorf("Failed! expected: %v, got: %v", mockMessage, gotMessage)
	} else {
		t.Logf("Success! expected: %v, got: %v", mockMessage, gotMessage)
	}
}

func TestClientManager_RegisterSubscriber(t *testing.T) {
	mockManager.Register <- mockClient.Client

	gotClient := <-mockManager.Register
	if gotClient != mockClient.Client {
		t.Errorf("Failed! expected: %v, got: %v", mockClient.Client, gotClient)
	} else {
		t.Logf("Success! expected: %v, got: %v", mockClient.Client, gotClient)
	}
}

func TestClientManager_UnregisterSubscriber(t *testing.T) {
	mockManager.Unregister <- mockClient.Client

	gotClient := <-mockManager.Unregister
	if gotClient != mockClient.Client {
		t.Errorf("Failed! expected: %v, got: %v", mockClient.Client, gotClient)
	} else {
		t.Logf("Success! expected: %v, got: %v", mockClient.Client, gotClient)
	}
}
