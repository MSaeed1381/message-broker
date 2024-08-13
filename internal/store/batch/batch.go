package batch

import (
	"context"
	"errors"
	"github.com/MSaeed1381/message-broker/internal/model"
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

type BulkInserter func(ctx context.Context, items []*Item)
type Handler struct {
	items        chan *Item
	bulkInsertFn BulkInserter
	config       Config
}

func NewBatchHandler(bulkInsert BulkInserter, config Config) *Handler {
	handler := &Handler{
		bulkInsertFn: bulkInsert,
		items:        make(chan *Item, 1),
		config:       config,
	}
	go handler.FlushBufferScheduler()

	return handler
}

func (h *Handler) FlushBufferScheduler() {
	buffer := make([]*Item, 0, h.config.BufferSize)
	ticker := time.NewTicker(h.config.FlushDuration)

	defer ticker.Stop() // TODO how to close ticker?

	flush := func() {
		if len(buffer) == 0 {
			return
		}

		h.bulkInsertFn(context.Background(), buffer)
		buffer = buffer[len(buffer):]
	}

	for {
		select {
		case item := <-h.items:
			buffer = append(buffer, item)
			if len(buffer) >= h.config.BufferSize {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}

func (h *Handler) AddAndWait(ctx context.Context, msg *model.Message) error {
	item := NewItem(msg)
	h.items <- item // write new item on a safe channel that flusher can communicate with that

	select {
	case <-ctx.Done():
		return errors.New("context canceled")
	case <-item.Done:
		if item.Message.BrokerMessage.Id == 0 {
			return errors.New("message saving failed")
		}
		return nil
	case <-time.After(h.config.MessageResponseTimeout):
		return errors.New("message response timeout")
	}
}
