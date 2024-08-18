package main

import (
	"github.com/MSaeed1381/message-broker/internal/broker"
	"github.com/MSaeed1381/message-broker/internal/store/batch"
	"github.com/MSaeed1381/message-broker/internal/store/cache"
	"github.com/MSaeed1381/message-broker/internal/store/postgres"
	"github.com/MSaeed1381/message-broker/internal/store/scylla"
	"time"
)

type StoreType int

const (
	InMemory StoreType = 0
	Postgres StoreType = 1
	ScyllaDB StoreType = 2
)

type Config struct {
	grpcAddr         string
	storeType        StoreType
	postgres         postgres.Config
	scylla           scylla.Config
	metricEnable     bool
	metricAddress    string
	profilerAddress  string
	cacheEnable      bool
	cache            cache.Config
	broker           broker.Config
	kubernetesEnable bool
}

func DefaultConfig() Config {
	return Config{
		grpcAddr:  "0.0.0.0:8000",
		storeType: ScyllaDB,
		postgres: postgres.Config{
			JdbcUri:        "postgres://postgres:postgres@localhost:5432/message_broker",
			MaxConnections: 1,
			MinConnections: 1,
			BatchConfig: batch.Config{
				BufferSize:             1024,
				FlushDuration:          time.Duration(1000) * time.Millisecond,
				MessageResponseTimeout: time.Duration(10) * time.Second,
			},
		},
		scylla: scylla.Config{
			Address:        "scylla",
			Keyspace:       "message_broker",
			NumConnections: 4,
			BatchConfig: batch.Config{
				BufferSize:             256,
				FlushDuration:          time.Duration(500) * time.Millisecond,
				MessageResponseTimeout: time.Duration(5) * time.Second,
			},
		},
		metricEnable:    true,
		metricAddress:   "0.0.0.0:5555",
		profilerAddress: "0.0.0.0:8080",
		cacheEnable:     false,
		cache: cache.Config{
			Address:  "cache:6379",
			Password: "",
			DBNumber: 0,
		},
		broker: broker.Config{
			ChannelBufferSize: 100,
		},
		kubernetesEnable: false,
	}
}
