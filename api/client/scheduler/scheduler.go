package scheduler

import (
	"context"
	"github.com/MSaeed1381/message-broker/api/client/option"
	"github.com/MSaeed1381/message-broker/api/client/worker"
	"github.com/MSaeed1381/message-broker/api/proto"
	"sync"
	"time"
)

type Scheduler struct {
	timeout    *time.Timer
	ticker     *time.Ticker
	clientPool worker.Client
}

func NewScheduler(client worker.Client) *Scheduler {
	return &Scheduler{clientPool: client}
}

func (s *Scheduler) Start(ctx context.Context, timeout, ticker time.Duration,
	job worker.HandleFunc, noRequests int, wg *sync.WaitGroup) {
	s.timeout = time.NewTimer(timeout * time.Second)
	s.ticker = time.NewTicker(ticker * time.Second)

	defer s.ticker.Stop()
	defer s.timeout.Stop()

	for {
		select {
		case <-ctx.Done():
			wg.Done()
			return
		case <-s.timeout.C: // time finishes
			wg.Done()
			return
		case <-s.ticker.C:
			for i := 0; i < noRequests; i++ {
				s.clientPool.SubmitJob(job)
			}
		}
	}
}

func (s *Scheduler) StartPublish(ctx context.Context, timeout, ticker time.Duration,
	request *proto.PublishRequest, noRequests int, wg *sync.WaitGroup) {
	job := worker.HandleFunc{
		Method:      option.Publish,
		Ctx:         ctx,
		Request:     request,
		ExpResponse: nil,
	}

	s.Start(ctx, timeout, ticker, job, noRequests, wg)
}

func (s *Scheduler) StartFetch(ctx context.Context, timeout, ticker time.Duration, request *proto.FetchRequest,
	noRequests int, response *proto.MessageResponse, wg *sync.WaitGroup) {
	job := worker.HandleFunc{
		Method:      option.Fetch,
		Ctx:         ctx,
		Request:     request,
		ExpResponse: response,
	}

	s.Start(ctx, timeout, ticker, job, noRequests, wg)
}
