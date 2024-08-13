package postgres

import "github.com/MSaeed1381/message-broker/internal/store/batch"

type Config struct {
	JdbcUri        string
	MaxConnections int
	MinConnections int
	BatchConfig    batch.Config
}
