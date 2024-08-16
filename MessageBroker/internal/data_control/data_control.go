package datacontrol

import (
	"therealbroker/pkg/broker"
)

type DataControl interface {
	SaveMessage(msg broker.Message) (string, error)
	RetriveMessage(id string) (broker.Message, error)
	ClearData() error
}
