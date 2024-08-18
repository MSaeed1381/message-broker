package cluster

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"github.com/MSaeed1381/message-broker/api/proto"
	"google.golang.org/grpc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log"
)

type Client struct{}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) hashSubjectToPod(subject string) string {
	pods, err := c.getPodIPs() // list of IPs
	if err != nil {
		log.Fatalf("Error getting pod IPs: %v", err)
	}

	hash := sha256.Sum256([]byte(subject))
	hashInt := binary.BigEndian.Uint64(hash[:8])
	podIndex := int(hashInt % uint64(len(pods)))
	return pods[podIndex]
}

func (c *Client) getPodIPs() ([]string, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	namespace := "default"
	labelSelector := "app=message-broker"

	pods, err := clientSet.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, err
	}

	var podIPs []string
	for _, pod := range pods.Items {
		podIPs = append(podIPs, pod.Status.PodIP)
	}

	return podIPs, nil
}

func (c *Client) PublishToAllPods(ctx context.Context, req *proto.PublishRequest) {
	podIPs, err := c.getPodIPs()
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

func (c *Client) PublishToTargetPod(ctx context.Context, request *proto.PublishRequest) {
	podIP := c.hashSubjectToPod(request.GetSubject())
	go func(podIP string) {
		address := fmt.Sprintf("%s:8000", podIP)

		conn, err := grpc.Dial(address, grpc.WithInsecure())
		if err != nil {
			log.Printf("Failed to connect to pod %s: %v", podIP, err)
			return
		}
		defer conn.Close()

		client := proto.NewBrokerClient(conn)

		_, err = client.Publish(ctx, request)
		if err != nil {
			log.Printf("Failed to publish message to pod %s: %v", podIP, err)
		} else {
			log.Printf("Successfully published message to pod %s", podIP)
		}
	}(podIP)
}
