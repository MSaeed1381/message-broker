package scylla

import "github.com/MSaeed1381/message-broker/internal/store/batch"

type Config struct {
	Address        string
	Keyspace       string
	NumConnections int
	BatchConfig    batch.Config
}
