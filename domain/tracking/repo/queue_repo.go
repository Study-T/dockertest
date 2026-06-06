package repo

import "context"

// QueueMessage represents a message from Redis queue.
type QueueMessage struct {
	RawData []byte
}

// QueueRepo defines interface for Redis queue operations.
type QueueRepo interface {
	// BRPop blocks and waits for messages from queue.
	BRPop(ctx context.Context, timeout int64) (*QueueMessage, error)

	// LLen returns the length of the queue.
	LLen(ctx context.Context) (int64, error)
}
