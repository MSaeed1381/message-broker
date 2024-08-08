package main

import (
	"context"
	"github.com/MSaeed1381/message-broker/api/server"
	"github.com/MSaeed1381/message-broker/internal/broker"
	"github.com/MSaeed1381/message-broker/internal/store"
	"github.com/MSaeed1381/message-broker/internal/store/memory"
	"github.com/MSaeed1381/message-broker/internal/store/postgres"
	"github.com/MSaeed1381/message-broker/pkg/metric"
	"github.com/prometheus/client_golang/prometheus"
)

// Main requirements:
// 1. All tests should be passed
// 2. Your logs should be accessible in Graylog
// 3. Basic prometheus.yaml metrics ( latency, throughput, etc. ) should be implemented
// 	  for every base functionality ( publish, subscribe etc. )

func main() {
	config := Config{
		grpcAddr:      "0.0.0.0:8000",
		storeType:     Postgres,
		postgresURI:   "postgres://postgres:postgres@postgres:5432/message_broker",
		metricEnable:  true,
		metricAddress: "0.0.0.0:5555",
	}

	// only for persist store in database (for in-memory data store is nil)
	var msgStore store.Message

	switch config.storeType {
	case Postgres:
		psql, err := postgres.NewPG(context.Background(), config.postgresURI) // connect to postgres
		if err != nil {
			panic(err)
		}
		defer psql.Close()
		msgStore = postgres.NewMessageInPostgres(*psql)
	case ScyllaDB: // TODO implement Scylla Database
		break
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

	brokerModule := broker.NewModule(topicStore)
	grpcServer := server.NewBrokerServer(brokerModule, prometheusController)
	grpcServer.Serve(config.grpcAddr)
}
