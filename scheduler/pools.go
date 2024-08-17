package scheduler

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	pb "github.com/pablovarg/distributed-task-scheduler/internal/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var ErrNoWorkers = errors.New("no available workers")

type WorkerPool struct {
	sync.RWMutex
	workers map[string]*Worker
	ids     []string
	counter int
}

type Worker struct {
	addr          string
	lastHeartbeat time.Time
	conn          *grpc.ClientConn
	client        pb.WorkerClient
}

func (pool *WorkerPool) handleHeartbeat(addr string) error {
	pool.Lock()
	defer pool.Unlock()

	worker, ok := pool.workers[addr]
	if !ok {
		conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return fmt.Errorf("could not create grpc client for worker [%w]", err)
		}

		pool.ids = append(pool.ids, addr)
		pool.workers[addr] = &Worker{
			addr:          addr,
			lastHeartbeat: time.Now(),
			conn:          conn,
			client:        pb.NewWorkerClient(conn),
		}
		return nil
	}

	worker.lastHeartbeat = time.Now()
	return nil
}

func (pool *WorkerPool) cleanWorkersContext(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			for _, worker := range pool.workers {
				worker.conn.Close()
			}
			return
		case <-time.After(1 * time.Second):
			pool.cleanWorkers()
		}
	}
}

func (pool *WorkerPool) cleanWorkers() {
	pool.Lock()
	defer pool.Unlock()

	newIDs := []string{}
	for addr, worker := range pool.workers {
		if time.Now().Sub(worker.lastHeartbeat) > 5*time.Second { // TODO: Get this value from .env
			delete(pool.workers, addr)
			continue
		}

		newIDs = append(newIDs, addr)
	}

	pool.ids = newIDs
	pool.counter = 0
}

func (pool *WorkerPool) nextWorker() (string, error) {
	pool.Lock()
	defer pool.Unlock()

	if len(pool.ids) == 0 {
		return "", ErrNoWorkers
	}

	selectedId := pool.ids[pool.counter%len(pool.ids)]
	pool.counter++

	return selectedId, nil
}
