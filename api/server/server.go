package server

import (
	"context"
	"github.com/MSaeed1381/message-broker/api/proto"
	"github.com/MSaeed1381/message-broker/pkg/broker"
	"google.golang.org/grpc"
	"log"
	"net"
	"time"
)

type BrokerServer struct {
	proto.UnimplementedBrokerServer
	service broker.Broker
}

func NewBrokerServer(service broker.Broker) *BrokerServer {
	return &BrokerServer{
		service: service,
	}
}

func (s *BrokerServer) Serve(addr string, service broker.Broker) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		panic("failed to listen: " + err.Error())
	}

	grpcServer := grpc.NewServer()
	proto.RegisterBrokerServer(grpcServer, NewBrokerServer(service))

	log.Printf("gRPC server listening at %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		panic("failed to serve: " + err.Error())
	}
}

func (s *BrokerServer) Publish(ctx context.Context, req *proto.PublishRequest) (*proto.PublishResponse, error) {
	messageID, err := s.service.Publish(ctx, req.GetSubject(), broker.Message{
		Body:       string(req.GetBody()),
		Expiration: time.Duration(req.GetExpirationSeconds()),
	})

	if err != nil {
		return nil, err
	}

	return &proto.PublishResponse{Id: int32(messageID)}, nil
}

func (s *BrokerServer) Subscribe(req *proto.SubscribeRequest, res proto.Broker_SubscribeServer) error {
	msgChannel, err := s.service.Subscribe(context.Background(), req.GetSubject())
	if err != nil {
		return err
	}

	for msg := range msgChannel {
		if err := res.Send(&proto.MessageResponse{Body: []byte(msg.Body)}); err != nil {
			return err
		}
	}
	return nil
}

func (s *BrokerServer) Fetch(ctx context.Context, req *proto.FetchRequest) (*proto.MessageResponse, error) {
	msg, err := s.service.Fetch(ctx, req.GetSubject(), int(req.GetId()))
	if err != nil {
		return nil, err
	}

	return &proto.MessageResponse{Body: []byte(msg.Body)}, nil
}
