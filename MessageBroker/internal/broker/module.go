package broker

import (
	"context"
	"sync"
	"therealbroker/pkg/broker"
	"time"
)

type Module struct {
	// TODO: Add required fields
	subscriptions  map[string][]chan broker.Message
	expirationTime map[int]time.Time
	message        map[int]broker.Message
	closed         bool
	bufferSize     int
	lock           sync.Mutex
}

func NewModule() broker.Broker {
	return &Module{
		subscriptions:  make(map[string][]chan broker.Message),
		expirationTime: make(map[int]time.Time),
		message:        make(map[int]broker.Message),
		closed:         false,
		bufferSize:     1000,
		lock:           sync.Mutex{},
	}
}

func (m *Module) Close() error {
	m.closed = true
	for _, subs := range m.subscriptions {
		for _, ch := range subs {
			close(ch)
		}
	}
	return nil
}

func (m *Module) Publish(ctx context.Context, subject string, msg broker.Message) (int, error) {
	if m.closed {
		return 0, broker.ErrUnavailable
	}

	m.lock.Lock()
	if _, ok := m.subscriptions[subject]; !ok {
		m.subscriptions[subject] = make([]chan broker.Message, 0)
	}

	for _, sub := range m.subscriptions[subject] {
		sub <- msg
	}

	m.expirationTime[msg.Id] = time.Now().Add(msg.Expiration)
	m.message[msg.Id] = msg
	m.lock.Unlock()
	return msg.Id, nil
}

func (m *Module) Subscribe(ctx context.Context, subject string) (<-chan broker.Message, error) {
	if m.closed {
		return nil, broker.ErrUnavailable
	}

	m.lock.Lock()
	if _, ok := m.subscriptions[subject]; !ok {
		m.subscriptions[subject] = make([]chan broker.Message, 0)
	}

	newsub := make(chan broker.Message, m.bufferSize)
	m.subscriptions[subject] = append(m.subscriptions[subject], newsub)
	m.lock.Unlock()
	return newsub, nil
}

func (m *Module) Fetch(ctx context.Context, subject string, id int) (broker.Message, error) {
	if m.closed {
		return broker.Message{}, broker.ErrUnavailable
	}
	if _, ok := m.message[id]; !ok {
		return broker.Message{}, broker.ErrInvalidID
	}
	if time.Now().After(m.expirationTime[id]) {
		return broker.Message{}, broker.ErrExpiredID
	}
	return m.message[id], nil
}
