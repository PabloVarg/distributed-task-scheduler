package worker

import (
	"context"
	"io"
	"log"
	"net"
	"os/exec"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/pablovarg/distributed-task-scheduler/internal/grpc"
	"github.com/pablovarg/distributed-task-scheduler/internal/task"
)

type WorkerConf struct {
	GRPCAddr      string
	WorkerAddr    string
	SchedulerAddr string
	Logger        *log.Logger
}

type Worker struct {
	ctx             context.Context
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

	w.ctx = ctx

	conn, err := grpc.NewClient(w.WorkerConf.SchedulerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		w.logger.Fatalln(err)
	}
	defer conn.Close()

	w.schedulerClient = pb.NewSchedulerClient(conn)
	go w.sendHeartbeats(ctx)
	go w.startGRPCServer(ctx)

	select {
	case <-ctx.Done():
		close(done)
	}
	return done
}

func (w *Worker) startGRPCServer(ctx context.Context) {
	lis, err := net.Listen("tcp", w.WorkerConf.GRPCAddr)
	if err != nil {
		w.logger.Fatalln(err)
	}

	server := grpc.NewServer()
	pb.RegisterWorkerServer(server, &WorkerServerImpl{
		Worker: w,
	})

	go func(ctx context.Context) {
		select {
		case <-ctx.Done():
			server.GracefulStop()
		}
	}(ctx)

	w.logger.Printf("Grpc server listening on %s", w.WorkerConf.GRPCAddr)
	if err := server.Serve(lis); err != nil {
		w.logger.Fatalln(err)
	}
}

func (w *Worker) sendHeartbeats(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(2 * time.Second): // TODO: Get from ENV
			w.logger.Printf("Sending heartbeat from %s", w.WorkerConf.WorkerAddr)
			w.schedulerClient.SendHeartbeat(ctx, &pb.Heartbeat{
				Address: w.WorkerConf.WorkerAddr,
			})
		}
	}
}

func (w *Worker) executeJob(task task.Task) {
	go func(ctx context.Context) {
		w.logger.Printf("executing %s", task.Command)

		w.schedulerClient.UpdateJobStatus(ctx, &pb.TaskStatus{
			ID:    int64(task.ID),
			State: pb.TaskState_PICKED,
		})

		out, err := exec.CommandContext(ctx, "sh", "-c", task.Command).Output()
		if err != nil {
			w.schedulerClient.UpdateJobStatus(ctx, &pb.TaskStatus{
				ID:    int64(task.ID),
				State: pb.TaskState_FAILED,
			})
			w.logger.Fatalln(err)
		}

		w.schedulerClient.UpdateJobStatus(ctx, &pb.TaskStatus{
			ID:    int64(task.ID),
			State: pb.TaskState_SUCCESS,
		})
		w.logger.Println(string(out))
	}(w.ctx)
}
