package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/pablovarg/distributed-task-scheduler/internal/env"
	"github.com/pablovarg/distributed-task-scheduler/scheduler"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	run(ctx, defaultLogger())
}

func run(ctx context.Context, logger *slog.Logger) {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	schedulerConf := readConf(logger)
	scheduler, err := scheduler.NewScheduler(schedulerConf)
	if err != nil {
		logger.Error(err.Error())
		panic(err)
	}

	logger.Info("starting scheduler")
	<-scheduler.Start(ctx)
	logger.Info("exiting scheduler")
}

func readConf(logger *slog.Logger) scheduler.SchedulerConf {
	dsn := env.GetRequiredEnvString("POSTGRES_DSN", logger)
	apiAddr := env.GetRequiredEnvString("API_ADDR", logger)
	grpcAddr := env.GetRequiredEnvString("GRPC_ADDR", logger)
	batchSize := env.GetEnvInt("BATCH_SIZE", 0, logger)
	workerDeadPeriod := env.GetRequiredEnvDuration("WORKER_DEAD_PERIOD", logger)
	pollInterval := env.GetRequiredEnvDuration("POLL_INTERVAL", logger)

	return scheduler.SchedulerConf{
		Addr:             apiAddr,
		GRPCAddr:         grpcAddr,
		DB_DSN:           dsn,
		Logger:           logger,
		PollInterval:     pollInterval,
		BatchSize:        batchSize,
		WorkerDeadPeriod: workerDeadPeriod,
	}
}

func defaultLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, nil))
}
