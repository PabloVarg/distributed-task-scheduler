package scheduler

import (
	"context"
	"time"

	pb "github.com/pablovarg/distributed-task-scheduler/internal/grpc"
)

type SchedulerServerImpl struct {
	*Scheduler
	pb.UnimplementedSchedulerServer
}

func (s *SchedulerServerImpl) SendHeartbeat(ctx context.Context, heartbeat *pb.Heartbeat) (*pb.Ok, error) {
	err := s.handleHeartbeat(heartbeat.GetAddress())
	if err != nil {
		return nil, err
	}

	return &pb.Ok{}, nil
}

func (s *SchedulerServerImpl) UpdateJobStatus(ctx context.Context, task *pb.TaskStatus) (*pb.Ok, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var err error
	switch task.GetState() {
	case pb.TaskState_PICKED:
		s.logger.Info("update task", "ID", task.GetID(), "status", "PICKED")
		err = s.taskModel.PickTask(ctx, int(task.GetID()))
	case pb.TaskState_SUCCESS:
		s.logger.Info("update task", "ID", task.GetID(), "status", "SUCCESS")
		err = s.taskModel.CompleteTask(ctx, int(task.GetID()))
	case pb.TaskState_FAILED:
		s.logger.Info("update task", "ID", task.GetID(), "status", "FAILED")
		err = s.taskModel.FailTask(ctx, int(task.GetID()))
	}

	if err != nil {
		return nil, err
	}

	return &pb.Ok{}, nil
}
