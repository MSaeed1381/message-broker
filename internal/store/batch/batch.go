package batch

import (
	"context"
	"github.com/MSaeed1381/message-broker/internal/model"
	"math"
	"time"
)

type Item struct {
	Message *model.Message
	Done    chan struct{}
}

func NewItem(message *model.Message) *Item {
	return &Item{
		Message: message,
		Done:    make(chan struct{}, 1),
	}
}

type BulkInserter func(ctx context.Context, messages []*model.Message)
type Handler struct {
	items         chan *Item
	bulkInsertFn  BulkInserter
	maxBufferSize int
}

func NewBatchHandler(bulkInsert BulkInserter, bufferSize int) *Handler {
	handler := &Handler{
		bulkInsertFn:  bulkInsert,
		items:         make(chan *Item, 1),
		maxBufferSize: bufferSize,
	}
	go handler.Resolve()

	return handler
}

func (h *Handler) Resolve() {
	buffer := make([]*Item, 0, h.maxBufferSize)
	ticker := time.NewTicker(5 * time.Second)

	flush := func() {
		if len(buffer) == 0 {
			return
		}

		tmp := make([]*model.Message, 0, len(buffer))
		for _, item := range buffer {
			tmp = append(tmp, item.Message)
		}

		maxIndex := int(math.Min(float64(len(buffer)), float64(h.maxBufferSize)))
		h.bulkInsertFn(context.Background(), tmp)

		for _, item := range buffer[:maxIndex] {
			close(item.Done)
		}
		buffer = buffer[maxIndex:]
	}

	for {
		select {
		case item := <-h.items:
			buffer = append(buffer, item)
			if len(buffer) >= h.maxBufferSize {
				flush()
			}

		case <-ticker.C:
			flush()
		}
	}
}

func (h *Handler) AddAndWait(ctx context.Context, msg *model.Message) {
	item := NewItem(msg)
	h.items <- item

	select {
	case <-ctx.Done():
		return
	case <-item.Done:
	}
}
