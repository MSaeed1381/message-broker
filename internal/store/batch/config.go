package batch

import "time"

type Config struct {
	BufferSize             int
	FlushDuration          time.Duration
	MessageResponseTimeout time.Duration
}
