package scheduler

import (
	"context"
	"sync"
	"time"
)

type WorkerPool struct {
	sync.RWMutex
	workers map[string]*Worker
}

type Worker struct {
	addr          string
	lastHeartbeat time.Time
}

func (pool *WorkerPool) handleHeartbeat(addr string) {
	pool.Lock()
	defer pool.Unlock()

	worker, ok := pool.workers[addr]
	if !ok {
		pool.workers[addr] = &Worker{
			addr:          addr,
			lastHeartbeat: time.Now(),
		}
		return
	}

	worker.lastHeartbeat = time.Now()
}

func (pool *WorkerPool) cleanWorkersContext(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(1 * time.Second):
			pool.cleanWorkers()
		}
	}
}

func (pool *WorkerPool) cleanWorkers() {
	pool.Lock()
	defer pool.Unlock()

	for addr, worker := range pool.workers {
		if time.Now().Sub(worker.lastHeartbeat) > 5*time.Second { // TODO: Get this value from .env
			delete(pool.workers, addr)
		}
	}
}
