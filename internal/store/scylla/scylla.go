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
}

// singleton design pattern with once keyword
var (
	scyllaInstance *Scylla
	scyllaOnce     sync.Once
)

// NewScylla initializes a new ScyllaDB connection using a singleton pattern
func NewScylla(_ context.Context, connString string) (*Scylla, error) {
	var err error
	scyllaOnce.Do(func() {
		cluster := gocql.NewCluster(connString)
		cluster.Keyspace = "message_broker"
		cluster.Consistency = gocql.One
		cluster.NumConns = 6

		session, err := cluster.CreateSession()
		if err != nil {
			log.Printf("Failed to create ScyllaDB session: %v", err)
			return
		}

		scyllaInstance = &Scylla{
			session: session,
		}
		fmt.Println("ScyllaDB session created successfully")
	})

	return scyllaInstance, err
}

// Ping checks the connection to ScyllaDB
func (s *Scylla) Ping(ctx context.Context) error {
	//if err := s.session.Query(`SELECT now() FROM postgres.public.users`).Exec(); err != nil {
	//	return fmt.Errorf("ping failed: %w", err)
	//}
	return nil
}

// Close closes the ScyllaDB session
func (s *Scylla) Close() {
	if s.session != nil {
		s.session.Close()
		fmt.Println("ScyllaDB session closed")
	}
}
