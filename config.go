package main

type StoreType int

const (
	InMemory StoreType = 0
	Postgres StoreType = 1
	ScyllaDB StoreType = 2
)

type Config struct {
	grpcAddr      string
	storeType     StoreType
	postgresURI   string
	metricEnable  bool
	metricAddress string
}
