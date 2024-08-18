package server

import (
	"context"
	"github.com/MSaeed1381/message-broker/api/proto"
	"github.com/MSaeed1381/message-broker/pkg/broker"
	"github.com/MSaeed1381/message-broker/pkg/metric"
	"google.golang.org/grpc"
	"log"
	"net"
	"time"
)

type BrokerServer struct {
	proto.UnimplementedBrokerServer
	service              broker.Broker
	prometheusController metric.Metric
}

func NewBrokerServer(service broker.Broker, pc metric.Metric) *BrokerServer {
	return &BrokerServer{
		service:              service,
		prometheusController: pc,
	}
}

func (s *BrokerServer) Serve(addr string) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		panic("failed to listen: " + err.Error())
	}

	grpcServer := grpc.NewServer()
	proto.RegisterBrokerServer(grpcServer, s)

	log.Printf("gRPC server listening at %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		panic("failed to serve: " + err.Error())
	}
}

func (s *BrokerServer) Publish(ctx context.Context, req *proto.PublishRequest) (*proto.PublishResponse, error) {
	start := time.Now()

	msg := &broker.Message{
		Body:       string(req.GetBody()),
		Expiration: time.Duration(float64(req.GetExpirationSeconds()) * float64(time.Second)),
	}

	messageID, err := s.service.Publish(ctx, req.GetSubject(), *msg)

	if err != nil {
		s.prometheusController.IncMethodCallCount(metric.Publish, metric.FAILURE)
		s.prometheusController.ObserveMethodDuration(metric.Publish, metric.FAILURE, time.Since(start).Seconds())
		return nil, err
	}

	s.prometheusController.IncMethodCallCount(metric.Publish, metric.SUCCESS)
	s.prometheusController.ObserveMethodDuration(metric.Publish, metric.SUCCESS, time.Since(start).Seconds())
	return &proto.PublishResponse{Id: messageID}, nil
}

func (s *BrokerServer) Subscribe(req *proto.SubscribeRequest, res proto.Broker_SubscribeServer) error {
	start := time.Now()

	msgChannel, err := s.service.Subscribe(res.Context(), req.GetSubject())
	if err != nil {
		s.prometheusController.IncMethodCallCount(metric.Subscribe, metric.FAILURE)
		s.prometheusController.ObserveMethodDuration(metric.Subscribe, metric.FAILURE, time.Since(start).Seconds())
		return err
	}

	s.prometheusController.IncMethodCallCount(metric.Subscribe, metric.SUCCESS)
	s.prometheusController.ObserveMethodDuration(metric.Subscribe, metric.SUCCESS, time.Since(start).Seconds())
	s.prometheusController.IncActiveSubscribers()

	func(ctx context.Context, msgChannel <-chan broker.Message) {
		for {
			select {
			case msg := <-msgChannel:
				if err := res.Send(&proto.MessageResponse{Body: []byte(msg.Body)}); err != nil {
					panic(err)
				}
			case <-ctx.Done():
				s.prometheusController.DecActiveSubscribers()
				log.Println("user cancelled the request")
				return
			}
		}
	}(res.Context(), msgChannel)

	return nil
}

func (s *BrokerServer) Fetch(ctx context.Context, req *proto.FetchRequest) (*proto.MessageResponse, error) {
	start := time.Now()

	msg, err := s.service.Fetch(ctx, req.GetSubject(), req.GetId())
	if err != nil {
		s.prometheusController.IncMethodCallCount(metric.Fetch, metric.FAILURE)
		s.prometheusController.ObserveMethodDuration(metric.Fetch, metric.FAILURE, time.Since(start).Seconds())
		return nil, err
	}

	s.prometheusController.IncMethodCallCount(metric.Fetch, metric.SUCCESS)
	s.prometheusController.ObserveMethodDuration(metric.Fetch, metric.SUCCESS, time.Since(start).Seconds())
	return &proto.MessageResponse{Body: []byte(msg.Body)}, nil
}
