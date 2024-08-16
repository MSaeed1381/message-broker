package main

import (
	"context"
	"github.com/MSaeed1381/message-broker/api/server"
	"github.com/MSaeed1381/message-broker/internal/broker"
	"github.com/MSaeed1381/message-broker/internal/store"
	"github.com/MSaeed1381/message-broker/internal/store/batch"
	"github.com/MSaeed1381/message-broker/internal/store/cache"
	"github.com/MSaeed1381/message-broker/internal/store/memory"
	"github.com/MSaeed1381/message-broker/internal/store/postgres"
	"github.com/MSaeed1381/message-broker/internal/store/scylla"
	"github.com/MSaeed1381/message-broker/pkg/metric"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"time"
)

// Main requirements:
// 1. All tests should be passed
// 2. Your logs should be accessible in Graylog
// 3. Basic prometheus.yaml metrics ( latency, throughput, etc. ) should be implemented
// 	  for every base functionality ( publish, subscribe etc. )

func main() {
	config := Config{
		grpcAddr:  "0.0.0.0:8000",
		storeType: ScyllaDB,
		postgres: postgres.Config{
			JdbcUri:        "postgres://postgres:postgres@localhost:5432/message_broker",
			MaxConnections: runtime.NumCPU() * 2,
			MinConnections: 2,
			BatchConfig: batch.Config{
				BufferSize:             2000,
				FlushDuration:          time.Duration(500) * time.Millisecond,
				MessageResponseTimeout: time.Duration(5) * time.Second,
			},
		},
		scylla: scylla.Config{
			Address:        "scylla",
			Keyspace:       "message_broker",
			NumConnections: runtime.NumCPU(),
			BatchConfig: batch.Config{
				BufferSize:             2000,
				FlushDuration:          time.Duration(500) * time.Millisecond,
				MessageResponseTimeout: time.Duration(5) * time.Second,
			},
		},
		metricEnable:    true,
		metricAddress:   "0.0.0.0:5555",
		profilerAddress: "0.0.0.0:8080",
		cache: cache.Config{
			Address:  "cache:6379",
			Password: "",
			DBNumber: 0,
		},
		broker: broker.Config{
			ChannelBufferSize: 100,
		},
	}

	// create a webserver for profiling
	go func() {
		err := http.ListenAndServe(config.profilerAddress, nil)
		if err != nil {
			return
		}
	}()

	// only for persist store in database (for in-memory data store is nil)
	var msgStore store.Message

	switch config.storeType {
	case Postgres:
		psql, err := postgres.NewPG(context.Background(), config.postgres) // connect to postgres
		if err != nil {
			panic(err)
		}
		defer psql.Close()
		msgStore = postgres.NewMessageInPostgres(*psql)
	case ScyllaDB:
		scyllaInstance, err := scylla.NewScylla(context.Background(), config.scylla)
		if err != nil {
			panic(err)
		}
		defer scyllaInstance.Close()
		msgStore = scylla.NewMessageInScylla(scyllaInstance)
	case InMemory:
		msgStore = nil
	default:
		panic("unknown store type")
	}

	topicStore := memory.NewTopicInMemory(msgStore)

	var prometheusController metric.Metric
	if config.metricEnable {
		reg := prometheus.NewRegistry()
		prometheusController = metric.NewPrometheusController(reg)
		go prometheusController.Serve(reg, config.metricAddress) // start prometheus controller socket on another port
	} else {
		prometheusController = &metric.NoImpl{}
	}

	brokerModule := broker.NewModule(topicStore, cache.NewRedisClient(config.cache), config.broker)
	grpcServer := server.NewBrokerServer(brokerModule, prometheusController)
	grpcServer.Serve(config.grpcAddr)
}
