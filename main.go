package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Tak1za/ivar/manager"
	"github.com/Tak1za/ivar/subscriber"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var clientManager *manager.ClientManager

func main() {
	r := gin.Default()
	clientManager = manager.NewManager()
	go clientManager.Start()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r.GET("/ws/:id", wsHandler)
	r.Run(":" + port)
}

func checkOrigin() func(r *http.Request) bool {
	return func(r *http.Request) bool {
		return true
	}
}

func wsHandler(context *gin.Context) {
	wsUpgrader.CheckOrigin = checkOrigin()
	wsUpgrader.EnableCompression = true
	groupId := context.Param("id")
	conn, err := wsUpgrader.Upgrade(context.Writer, context.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade to websocket: %+v", err)
		return
	}

	newSubscriber := subscriber.NewSubscriber(conn, groupId)

	clientManager.RegisterSubscriber(newSubscriber.Client)

	go newSubscriber.Read(clientManager)
	go newSubscriber.Write(clientManager)
}
