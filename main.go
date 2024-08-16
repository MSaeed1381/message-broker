package main

import (
	"context"
	"github.com/MSaeed1381/message-broker/api/server"
	"github.com/MSaeed1381/message-broker/internal/broker"
	"github.com/MSaeed1381/message-broker/internal/store"
	"github.com/MSaeed1381/message-broker/internal/store/cache"
	"github.com/MSaeed1381/message-broker/internal/store/memory"
	"github.com/MSaeed1381/message-broker/internal/store/postgres"
	"github.com/MSaeed1381/message-broker/internal/store/scylla"
	"github.com/MSaeed1381/message-broker/pkg/metric"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	_ "net/http/pprof"
)

// Main requirements:
// 1. All tests should be passed
// 2. Your logs should be accessible in Graylog
// 3. Basic prometheus.yaml metrics ( latency, throughput, etc. ) should be implemented
// 	  for every base functionality ( publish, subscribe etc. )

func main() {
	config := DefaultConfig()
	go initProfiler(config)                                                 // create a webserver for profiling
	msgStore := initDataStore(config)                                       // only for persist store in database (for in-memory data store is nil)
	cacheStore := initCacheMemory(config)                                   // create cache store
	topicStore := memory.NewTopicInMemory(msgStore)                         // create new topic store
	prometheusController := initPrometheus(config)                          // define metric
	brokerModule := broker.NewModule(topicStore, cacheStore, config.broker) // create new broker module
	grpcServer := server.NewBrokerServer(brokerModule, prometheusController)
	grpcServer.Serve(config.grpcAddr)
}

func initDataStore(config Config) store.Message {
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

	return msgStore
}

func initCacheMemory(config Config) cache.Cache {
	var cacheStore cache.Cache
	if config.cacheEnable {
		cacheStore = cache.NewRedisClient(config.cache)
	} else {
		cacheStore = cache.NewNoImpl()
	}

	return cacheStore
}

func initPrometheus(config Config) metric.Metric {
	var prometheusController metric.Metric
	if config.metricEnable {
		reg := prometheus.NewRegistry() // create new registry for gRPC metrics
		prometheusController = metric.NewPrometheusController(reg)
		go prometheusController.Serve(reg, config.metricAddress) // start prometheus controller socket on another port
	} else {
		prometheusController = &metric.NoImpl{}
	}

	return prometheusController
}

func initProfiler(config Config) {
	err := http.ListenAndServe(config.profilerAddress, nil)
	if err != nil {
		return
	}
}
