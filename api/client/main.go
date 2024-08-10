package main

import (
	"context"
	"fmt"
	"github.com/MSaeed1381/message-broker/api/client/option"
	"github.com/MSaeed1381/message-broker/api/client/report"
	"github.com/MSaeed1381/message-broker/api/client/scheduler"
	"github.com/MSaeed1381/message-broker/api/client/worker"
	"github.com/MSaeed1381/message-broker/api/proto"
	"sync"
	"time"
)

func main() {
	// Override this Section to change the scenario
	opt := option.Option{
		Host: "localhost:8000",
		Scenario: option.Scenario{
			Executor:        "constant-arrival-rate",
			Rate:            10000,
			Timeunit:        1,
			Duration:        60,
			PreAllocatedVUs: 20,
			MaxVUs:          200,
		},
	}

	reportSummary := report.NewReport()
	client := worker.NewClientImpl(worker.Address(opt.Host), time.Duration(opt.Scenario.Duration), reportSummary)

	const NumberOfConnections = 30
	err := client.CreateClientPool(NumberOfConnections, opt.Scenario.Rate/10)
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup

	sch := scheduler.NewScheduler(client)

	wg.Add(1)
	publishRequest := &proto.PublishRequest{
		Subject:           "ali",
		Body:              []byte("hello_world"),
		ExpirationSeconds: 60,
	}
	go sch.StartPublish(context.Background(), time.Duration(opt.Scenario.Duration),
		time.Duration(opt.Scenario.Timeunit), publishRequest, opt.Scenario.Rate, &wg)

	//wg.Add(1)
	//fetchRequest := &proto.FetchRequest{
	//	Subject: "ali",
	//	Id:      517727,
	//}
	//
	//fetchResponse := &proto.MessageResponse{
	//	Body: []byte("saeed"),
	//}
	//
	//go sch.StartFetch(context.Background(), time.Duration(opt.Scenario.Duration),
	//	time.Duration(opt.Scenario.Timeunit), fetchRequest, opt.Scenario.Rate, fetchResponse, &wg)

	wg.Wait()

	fmt.Println(reportSummary)
}
