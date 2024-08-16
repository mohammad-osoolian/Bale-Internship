package datacontrol

import (
	"fmt"
	"sync"
	"therealbroker/pkg/broker"
	"time"
)

type DataMemory struct {
	DataControl
	expirationTime map[string]time.Time
	message        map[string]broker.Message
	messageId      int
	lock           sync.Mutex
}

func NewDataMemory() *DataMemory {
	return &DataMemory{
		expirationTime: make(map[string]time.Time),
		message:        make(map[string]broker.Message),
		messageId:      0,
	}
}

func (dm *DataMemory) ClearData() error {
	dm.lock.Lock()
	for k := range dm.expirationTime {
		delete(dm.expirationTime, k)
	}

	for k := range dm.message {
		delete(dm.message, k)
	}

	dm.messageId = 0
	dm.lock.Unlock()
	return nil
}

func (dm *DataMemory) SaveMessage(msg broker.Message) (string, error) {
	dm.lock.Lock()
	msg.Id = fmt.Sprintf("%v", dm.messageId)
	dm.messageId++

	// if dm.IdExists(msg.Id) {
	// 	return msg.Id, broker.ErrAlreadyExistID
	// }

	dm.expirationTime[msg.Id] = time.Now().Add(msg.Expiration)
	dm.message[msg.Id] = msg
	dm.lock.Unlock()
	return msg.Id, nil
}

func (dm *DataMemory) RetriveMessage(id string) (broker.Message, error) {
	dm.lock.Lock()
	if _, ok := dm.message[id]; !ok {
		return broker.Message{}, broker.ErrInvalidID
	}
	if time.Now().After(dm.expirationTime[id]) {
		return broker.Message{}, broker.ErrExpiredID
	}
	msg := dm.message[id]
	dm.lock.Unlock()
	return msg, nil
}

func (dm *DataMemory) IdExists(id string) bool {
	dm.lock.Lock()
	_, ok := dm.message[id]
	dm.lock.Unlock()
	return ok
}
