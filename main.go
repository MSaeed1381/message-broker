package main

import (
	"github.com/MSaeed1381/message-broker/api/server"
	"github.com/MSaeed1381/message-broker/internal/broker"
)

// Main requirements:
// 1. All tests should be passed
// 2. Your logs should be accessible in Graylog
// 3. Basic prometheus metrics ( latency, throughput, etc. ) should be implemented
// 	  for every base functionality ( publish, subscribe etc. )

func main() {
	config := Config{
		grpcAddr: "0.0.0.0:8000",
	}

	brokerModule := broker.NewModule()
	grpcServer := server.NewBrokerServer(brokerModule)

	grpcServer.Serve(config.grpcAddr)

}
