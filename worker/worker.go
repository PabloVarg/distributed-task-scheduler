package worker

import (
	"context"
	"fmt"
	"io"
	"log/slog"
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
	Logger        *slog.Logger
}

type Worker struct {
	ctx             context.Context
	schedulerClient pb.SchedulerClient
	logger          *slog.Logger
	WorkerConf
}

func NewWorker(conf WorkerConf) *Worker {
	assignedLogger := conf.Logger
	if assignedLogger == nil {
		assignedLogger = slog.New(slog.NewJSONHandler(io.Discard, nil))
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
		w.logger.Error(err.Error())
		panic(err)
	}

	w.schedulerClient = pb.NewSchedulerClient(conn)
	go w.sendHeartbeats(ctx)
	go w.startGRPCServer(ctx)

	go func(ctx context.Context) {
		select {
		case <-ctx.Done():
			conn.Close()
			close(done)
		}
	}(ctx)
	return done
}

func (w *Worker) startGRPCServer(ctx context.Context) {
	lis, err := net.Listen("tcp", w.WorkerConf.GRPCAddr)
	if err != nil {
		w.logger.Error(err.Error())
		panic(err)
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

	w.logger.Info(fmt.Sprintf("grpc server listening on %s", w.WorkerConf.GRPCAddr))
	if err := server.Serve(lis); err != nil {
		w.logger.Error(err.Error())
		panic(err)
	}
}

func (w *Worker) sendHeartbeats(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second) // TODO: Get from ENV
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.logger.Info("Sending heartbeat", "worker", w.WorkerConf.WorkerAddr)
			_, err := w.schedulerClient.SendHeartbeat(ctx, &pb.Heartbeat{
				Address: w.WorkerConf.WorkerAddr,
			})
			if err != nil {
				w.logger.Error(err.Error())
			}
		}
	}
}

func (w *Worker) executeJob(task task.Task) {
	go func(ctx context.Context) {
		w.logger.Info("executing", "command", task.Command)

		_, err := w.schedulerClient.UpdateJobStatus(ctx, &pb.TaskStatus{
			ID:    int64(task.ID),
			State: pb.TaskState_PICKED,
		})
		if err != nil {
			w.logger.Error("error picking task", "ID", task.ID, "error", err)
			return
		}

		out, err := exec.CommandContext(ctx, "sh", "-c", task.Command).Output()
		if err != nil {
			w.schedulerClient.UpdateJobStatus(ctx, &pb.TaskStatus{
				ID:    int64(task.ID),
				State: pb.TaskState_FAILED,
			})
			w.logger.Error(err.Error())
			panic(err)
		}

		_, err = w.schedulerClient.UpdateJobStatus(ctx, &pb.TaskStatus{
			ID:    int64(task.ID),
			State: pb.TaskState_SUCCESS,
		})
		if err != nil {
			w.logger.Error("error marking task as successful", "ID", task.ID)
			return
		}

		w.logger.Info("executed command", "output", string(out))
	}(w.ctx)
}
