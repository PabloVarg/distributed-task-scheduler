package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/pablovarg/distributed-task-scheduler/internal/env"
	"github.com/pablovarg/distributed-task-scheduler/worker"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	run(ctx, defaultLogger())
}

func run(ctx context.Context, logger *slog.Logger) {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	conf := readConf(logger)
	<-worker.NewWorker(conf).Start(ctx)
}

func readConf(logger *slog.Logger) worker.WorkerConf {
	schedulerAddr := env.GetRequiredEnvString("SCHEDULER_ADDR", logger)
	workerAddr := env.GetRequiredEnvString("WORKER_ADDR", logger)
	grpcAddr := env.GetRequiredEnvString("GRPC_ADDR", logger)
	heartbeatInterval := env.GetEnvDuration("HEARTBEAT_INTERVAL", "0s", logger)

	if heartbeatInterval <= 0 {
		logger.Warn("heartbeats will not be sent")
	}

	return worker.WorkerConf{
		GRPCAddr:          grpcAddr,
		WorkerAddr:        workerAddr,
		SchedulerAddr:     schedulerAddr,
		Logger:            logger,
		HeartbeatInterval: heartbeatInterval,
	}
}

func defaultLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, nil))
}
