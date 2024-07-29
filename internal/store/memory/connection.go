package memory

import (
	"context"
	"github.com/MSaeed1381/message-broker/internal/model"
	"sync"
)

type ConnectionInMemory struct {
	Connections []*model.Connection
	mutex       sync.Mutex
}

func NewConnectionInMemory() *ConnectionInMemory {
	return &ConnectionInMemory{
		Connections: make([]*model.Connection, 0),
		mutex:       sync.Mutex{},
	}
}

func (c *ConnectionInMemory) Save(ctx context.Context, connection *model.Connection) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.Connections = append(c.Connections, connection)

	return nil
}

func (c *ConnectionInMemory) GetAllConnections(ctx context.Context) ([]*model.Connection, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.Connections, nil
}
