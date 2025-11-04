package server

import (
	"log"
	"sync"
	"time"

	pb "github.com/nhulston/logstream/proto/gen"
)

type Batcher struct {
	maxSize   int
	maxWait   time.Duration
	flushFunc func([]*pb.LogEntry) error
	logs      []*pb.LogEntry
	mu        sync.Mutex
	ticker    *time.Ticker
	stopCh    chan struct{}
	doneCh    chan struct{}
}

func NewBatcher(maxSize int, maxWait time.Duration, flushFunc func([]*pb.LogEntry) error) *Batcher {
	return &Batcher{
		maxSize:   maxSize,
		maxWait:   maxWait,
		flushFunc: flushFunc,
		logs:      make([]*pb.LogEntry, 0, maxSize),
		stopCh:    make(chan struct{}),
		doneCh:    make(chan struct{}),
	}
}

func (b *Batcher) Start() {
	b.ticker = time.NewTicker(b.maxWait)
	defer close(b.doneCh)

	for {
		select {
		case <-b.ticker.C:
			b.flush()
		case <-b.stopCh:
			b.flush()
			return
		}
	}
}

func (b *Batcher) Add(log *pb.LogEntry) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.logs = append(b.logs, log)

	if len(b.logs) >= b.maxSize {
		b.flushLocked()
	}
}

func (b *Batcher) flush() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.flushLocked()
}

func (b *Batcher) flushLocked() {
	if len(b.logs) == 0 {
		return
	}

	if err := b.flushFunc(b.logs); err != nil {
		log.Printf("Failed to flush batch: %v", err)
	}

	b.logs = make([]*pb.LogEntry, 0, b.maxSize)
}

func (b *Batcher) Stop() {
	close(b.stopCh)
	<-b.doneCh
	b.ticker.Stop()
}
