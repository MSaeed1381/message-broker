package worker

import (
	"context"
	"errors"
	"github.com/MSaeed1381/message-broker/api/client/option"
	"github.com/MSaeed1381/message-broker/api/client/report"
	"github.com/MSaeed1381/message-broker/api/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"time"
)

type HandleFunc struct {
	Method      option.Method
	Ctx         context.Context
	Request     interface{}
	ExpResponse interface{}
}

type Client interface {
	CreateClientPool(noClients int, noWorkerInClient int) error
	SubmitJob(job HandleFunc)
}

type Address string

type ClientImpl struct {
	addr    Address
	timeout time.Duration
	channel chan HandleFunc
	clients []proto.BrokerClient
	report  *report.Report
}

func NewClientImpl(addr Address, timeout time.Duration, report *report.Report) *ClientImpl {
	return &ClientImpl{addr: addr,
		clients: make([]proto.BrokerClient, 0),
		timeout: timeout,
		channel: make(chan HandleFunc, 1000000),
		report:  report,
	}
}

func (p *ClientImpl) CreateClientPool(noClients int, noWorkerInClient int) error {
	// open many connections (NoConnections) to connect to the grpc server and create a gRPC client with Connection
	for i := 0; i < noClients; i++ {
		conn, err := grpc.Dial(string(p.addr), grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return errors.New("all clients failed to connect to the message broker server")
		}

		// adds new client to list
		p.clients = append(p.clients, proto.NewBrokerClient(conn))

		// after duration time connection must be close
		//go func(conn *grpc.ClientConn) {
		//	<-time.After(p.timeout * time.Second)
		//
		//	fmt.Println("client timeout. closing the connection...")
		//	err := conn.Close()
		//	if err != nil {
		//		log.Printf("close connection error: %v", err)
		//	}
		//}(conn)
	}

	for _, client := range p.clients {
		for i := 0; i < noWorkerInClient; i++ {
			go p.process(&client)
		}
	}

	return nil
}

func (p *ClientImpl) SubmitJob(job HandleFunc) {
	p.channel <- job
}

func (p *ClientImpl) process(brokerClient *proto.BrokerClient) {
	for job := range p.channel {
		switch job.Method {
		case option.Publish:
			publishId, err := (*brokerClient).Publish(job.Ctx, job.Request.(*proto.PublishRequest))
			if err != nil || publishId == nil {
				p.report.PublishFailure.Inc()
			} else {
				p.report.PublishSuccess.Inc()
			}
		case option.Fetch:
			fetchResponse, err := (*brokerClient).Fetch(job.Ctx, job.Request.(*proto.FetchRequest))
			if err != nil || fetchResponse == nil ||
				string(fetchResponse.Body) != string(job.ExpResponse.(*proto.MessageResponse).Body) {
				p.report.FetchFailure.Inc()
			} else {
				p.report.FetchSuccess.Inc()
			}
		}
	}
}
