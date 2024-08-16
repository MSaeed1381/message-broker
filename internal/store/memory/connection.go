package memory

import (
	"context"
	"github.com/MSaeed1381/message-broker/internal/model"
	"sync"
	"time"
)

type ConnectionInMemory struct {
	Connections []*model.Connection
	lsMutex     sync.Mutex
}

func NewConnectionInMemory() *ConnectionInMemory {
	return &ConnectionInMemory{
		Connections: make([]*model.Connection, 0),
		lsMutex:     sync.Mutex{},
	}
}

func (c *ConnectionInMemory) Save(_ context.Context, connection *model.Connection) error {
	c.lsMutex.Lock()
	defer c.lsMutex.Unlock()
	c.Connections = append(c.Connections, connection)

	return nil
}

func (c *ConnectionInMemory) GetAllConnections(_ context.Context) ([]*model.Connection, error) {
	c.lsMutex.Lock()
	defer c.lsMutex.Unlock()

	return c.Connections, nil
}

func (c *ConnectionInMemory) SendMessageToSubscribers(_ context.Context, msg *model.Message) error {
	var wg sync.WaitGroup
	c.lsMutex.Lock()
	for _, connection := range c.Connections {
		wg.Add(1)
		go func(connection *model.Connection) {
			defer wg.Done()
			select {
			case connection.Channel <- *msg.BrokerMessage:
				break
			case <-time.After(time.Second): // timeout -> drop the messages if subscribers don't get the messages
			}
		}(connection)
	}

	go func() {
		wg.Wait()
		c.lsMutex.Unlock()
	}()

	return nil
}
