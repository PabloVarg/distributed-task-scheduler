package scheduler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	pb "github.com/pablovarg/distributed-task-scheduler/internal/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	ErrNoWorkers         = errors.New("no available workers")
	ErrNonExistingWorker = errors.New("the selected worker does not exist")
)

type WorkerPool struct {
	sync.RWMutex
	workers          map[string]*Worker
	ids              []string
	counter          int
	logger           *slog.Logger
	workerDeadPeriod time.Duration
	staticWorkerAddr string
}

type Worker struct {
	addr          string
	lastHeartbeat time.Time
	conn          *grpc.ClientConn
	client        pb.WorkerClient
}

func (pool *WorkerPool) Start(ctx context.Context) {
	if pool.staticWorkerAddr != "" {
		conn, err := grpc.NewClient(pool.staticWorkerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			err := fmt.Errorf("could not create grpc client for static worker")
			pool.logger.Error(err.Error())
			panic(err)
		}

		pool.ids = append(pool.ids, pool.staticWorkerAddr)
		pool.workers[pool.staticWorkerAddr] = &Worker{
			addr:   pool.staticWorkerAddr,
			conn:   conn,
			client: pb.NewWorkerClient(conn),
		}

		pool.logger.Warn("static worker detected, all tasks will be sent to the same address", "worker", pool.staticWorkerAddr)
	}

	go pool.cleanWorkersContext(ctx)
}

func (pool *WorkerPool) handleHeartbeat(addr string) error {
	if pool.staticWorkerAddr != "" {
		return nil
	}
	pool.Lock()
	defer pool.Unlock()

	worker, ok := pool.workers[addr]
	if !ok {
		conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return fmt.Errorf("could not create grpc client for worker [%w]", err)
		}

		pool.logger.Info("detected new worker", "addr", addr)

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
	if pool.workerDeadPeriod <= 0 {
		return
	}

	ticker := time.NewTicker(pool.workerDeadPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			for _, worker := range pool.workers {
				worker.conn.Close()
			}
			return
		case <-ticker.C:
			if pool.staticWorkerAddr != "" {
				continue
			}

			pool.cleanWorkers()
		}
	}
}

func (pool *WorkerPool) cleanWorkers() {
	pool.Lock()
	defer pool.Unlock()

	newIDs := []string{}
	for addr, worker := range pool.workers {
		if time.Now().Sub(worker.lastHeartbeat) >= pool.workerDeadPeriod {
			pool.logger.Warn("lost contact with worker", "addr", addr)
			delete(pool.workers, addr)
			continue
		}

		newIDs = append(newIDs, addr)
	}

	pool.ids = newIDs
	pool.counter = 0
}

func (pool *WorkerPool) nextWorker() (string, error) {
	if pool.staticWorkerAddr != "" {
		return pool.staticWorkerAddr, nil
	}

	pool.Lock()
	defer pool.Unlock()

	if len(pool.ids) == 0 {
		return "", ErrNoWorkers
	}

	selectedId := pool.ids[pool.counter%len(pool.ids)]
	pool.counter++

	return selectedId, nil
}
