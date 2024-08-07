package worker

import (
	"context"
	"io"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/pablovarg/distributed-task-scheduler/internal/grpc"
)

type WorkerConf struct {
	WorkerAddr    string
	SchedulerAddr string
	Logger        *log.Logger
}

type Worker struct {
	schedulerClient pb.SchedulerClient
	logger          *log.Logger
	WorkerConf
}

func NewWorker(conf WorkerConf) *Worker {
	assignedLogger := conf.Logger
	if assignedLogger == nil {
		assignedLogger = log.New(io.Discard, "", 0)
	}

	worker := &Worker{
		WorkerConf: conf,
		logger:     assignedLogger,
	}
	return worker
}

func (w *Worker) Start(ctx context.Context) <-chan any {
	done := make(chan any)

	conn, err := grpc.NewClient(w.WorkerConf.SchedulerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		w.logger.Fatalln(err)
	}
	defer conn.Close()

	w.schedulerClient = pb.NewSchedulerClient(conn)
	go w.sendHeartbeats(ctx)

	select {
	case <-ctx.Done():
		close(done)
	}
	return done
}

func (w *Worker) sendHeartbeats(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(10 * time.Second): // TODO: Get from ENV
			w.logger.Printf("Sending heartbeat from %s", w.WorkerConf.WorkerAddr)
			w.schedulerClient.SendHeartbeat(ctx, &pb.Heartbeat{
				Address: w.WorkerConf.WorkerAddr,
			})
		}
	}
}
