package scylla

import (
	"context"
	"fmt"
	"github.com/gocql/gocql"
	"log"
	"sync"
)

type Scylla struct {
	session *gocql.Session
	config  Config
}

// singleton design pattern with once keyword
var (
	scyllaInstance *Scylla
	scyllaOnce     sync.Once
)

// NewScylla initializes a new ScyllaDB connection using a singleton pattern
func NewScylla(_ context.Context, conf Config) (*Scylla, error) {
	var err error
	scyllaOnce.Do(func() {
		cluster := gocql.NewCluster(conf.Address)
		cluster.Keyspace = conf.Keyspace
		cluster.Consistency = gocql.One
		cluster.NumConns = conf.NumConnections

		session, err := cluster.CreateSession()
		if err != nil {
			log.Printf("Failed to create ScyllaDB session: %v", err)
			return
		}

		scyllaInstance = &Scylla{
			session: session,
			config:  conf,
		}
		fmt.Println("ScyllaDB session created successfully")
	})

	return scyllaInstance, err
}

// Ping checks the connection to ScyllaDB
func (s *Scylla) Ping(_ context.Context) error {
	if err := s.session.Query(`SELECT count(*) FROM message_broker.message`).Exec(); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}
	return nil
}

// Close closes the ScyllaDB session
func (s *Scylla) Close() {
	if s.session != nil {
		s.session.Close()
		fmt.Println("ScyllaDB session closed")
	}
}
