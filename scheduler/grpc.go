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
	s.logger.Printf("Receiving heartbeat from %s", heartbeat.GetAddress())
	err := s.handleHeartbeat(heartbeat.GetAddress())
	if err != nil {
		return nil, err
	}

	return &pb.Ok{}, nil
}
}
