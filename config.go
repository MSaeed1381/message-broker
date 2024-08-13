package main

import (
	"github.com/MSaeed1381/message-broker/internal/store/postgres"
	"github.com/MSaeed1381/message-broker/internal/store/scylla"
)

type StoreType int

const (
	InMemory StoreType = 0
	Postgres StoreType = 1
	ScyllaDB StoreType = 2
)

type Config struct {
	grpcAddr        string
	storeType       StoreType
	pgConfig        postgres.Config
	scyllaConfig    scylla.Config
	metricEnable    bool
	metricAddress   string
	profilerAddress string
}
