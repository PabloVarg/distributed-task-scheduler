package scheduler

import (
	"context"

	pb "github.com/pablovarg/distributed-task-scheduler/internal/grpc"
)

type SchedulerServerImpl struct {
	*Scheduler
	pb.UnimplementedSchedulerServer
}

func (s *SchedulerServerImpl) SendHeartbeat(ctx context.Context, heartbeat *pb.Heartbeat) (*pb.Ok, error) {
	s.Scheduler.logger.Println("Receiving heartbeat")

	return &pb.Ok{
		Success: true,
	}, nil
}
