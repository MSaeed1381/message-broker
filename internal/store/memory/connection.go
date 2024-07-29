package memory

import (
	"sync"
	"therealbroker/internal/model"
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

func (c *ConnectionInMemory) AddConnection(connection *model.Connection) error {
	c.mutex.Lock()
	c.Connections = append(c.Connections, connection)
	c.mutex.Unlock()

	return nil
}
