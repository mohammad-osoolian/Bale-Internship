package datacontrol

import (
	"therealbroker/pkg/broker"
)

type DataControl interface {
	SaveMessage(msg broker.Message) (int, error)
	RetriveMessage(id int) (broker.Message, error)
	IdExists(id int) bool
	ClearData() error
}
