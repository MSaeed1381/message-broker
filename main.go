package main

import (
	"github.com/MSaeed1381/message-broker/api/server"
	"github.com/MSaeed1381/message-broker/internal/broker"
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
		grpcAddr: "0.0.0.0:8000",
	}

	brokerModule := broker.NewModule()

	reg := prometheus.NewRegistry()
	prometheusController := metric.NewPrometheusController(reg)
	go prometheusController.Serve(reg)

	grpcServer := server.NewBrokerServer(brokerModule, prometheusController)
	grpcServer.Serve(config.grpcAddr)
}
