package imanager

import (
	"github.com/Tak1za/ivar/models"
)

type IManager interface {
	BroadcastMessage(message models.Message)
	UnregisterSubscriber(client *models.Client)
}
