package broker

import (
	"context"
	"sync"
	datacontrol "therealbroker/internal/data_control"
	"therealbroker/pkg/broker"
)

type Module struct {
	// TODO: Add required fields
	subscriptions map[string][]chan broker.Message
	data          datacontrol.DataControl
	closed        bool
	bufferSize    int
	lock          sync.Mutex
}

func NewModule(data datacontrol.DataControl) broker.Broker {
	return &Module{
		subscriptions: make(map[string][]chan broker.Message),
		data:          data,
		closed:        false,
		bufferSize:    1000,
		lock:          sync.Mutex{},
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
	m.lock.Unlock()

	id, err := m.data.SaveMessage(msg)
	return id, err
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
	msg, err := m.data.RetriveMessage(id)
	return msg, err
}
