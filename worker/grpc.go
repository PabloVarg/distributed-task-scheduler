package worker

import (
	"context"

	pb "github.com/pablovarg/distributed-task-scheduler/internal/grpc"
	"github.com/pablovarg/distributed-task-scheduler/internal/task"
)

type WorkerServerImpl struct {
	*Worker
	pb.UnimplementedWorkerServer
}

func (w *WorkerServerImpl) ExecuteJob(ctx context.Context, sentTask *pb.Task) (*pb.Ok, error) {
	w.logger.Info("received task", "ID", sentTask.GetID())
	w.Worker.executeJob(task.Task{
		ID:      int(sentTask.GetID()),
		Command: sentTask.GetCommand(),
	})

	return &pb.Ok{}, nil
}
