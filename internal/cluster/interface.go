package cluster

import (
	"context"
	"fmt"
	"github.com/MSaeed1381/message-broker/api/proto"
	"google.golang.org/grpc"
	"log"
)

type KubeClient interface {
	getPodIPs() ([]string, error)
	PublishToAllPods(ctx context.Context, req *proto.PublishRequest)
	PublishToTargetPod(ctx context.Context, request *proto.PublishRequest)
	hashSubjectToPod(subject string) string
}

type NoImpl struct{}

func NewNoImpl() *NoImpl {
	return &NoImpl{}
}

func (n *NoImpl) getPodIPs() ([]string, error) {
	return []string{"0.0.0.0"}, nil
}

func (n *NoImpl) PublishToAllPods(ctx context.Context, req *proto.PublishRequest) {
	podIPs, err := n.getPodIPs()
	if err != nil {
		log.Fatalf("Error fetching pod IPs: %v", err)
	}

	for _, podIP := range podIPs {
		go func(ctx context.Context, podIP string) {
			address := fmt.Sprintf("%s:8000", podIP)

			conn, err := grpc.Dial(address, grpc.WithInsecure())
			if err != nil {
				log.Printf("Failed to connect to pod %s: %v", podIP, err)
				return
			}
			defer conn.Close()

			client := proto.NewBrokerClient(conn)
			_, err = client.Publish(ctx, req)
			if err != nil {
				log.Printf("Failed to publish message to pod %s: %v", podIP, err)
			} else {
				log.Printf("Successfully published message to pod %s", podIP)
			}
		}(ctx, podIP)
	}
}

func (n *NoImpl) PublishToTargetPod(_ context.Context, _ *proto.PublishRequest) {}
func (n *NoImpl) hashSubjectToPod(_ string) string                              { return "" }
