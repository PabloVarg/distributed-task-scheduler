package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"time"

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
	schedulerAddr, ok := os.LookupEnv("SCHEDULER_ADDR")
	if !ok {
		err := fmt.Errorf("scheduler address not found")

		logger.Error(err.Error())
		panic(err)
	}
	workerAddr, ok := os.LookupEnv("WORKER_ADDR")
	if !ok {
		err := fmt.Errorf("worker address not found")

		logger.Error(err.Error())
		panic(err)
	}
	grpcAddr, ok := os.LookupEnv("GRPC_ADDR")
	if !ok {
		err := fmt.Errorf("grpc address not found")

		logger.Error(err.Error())
		panic(err)
	}

	heartbeatInterval, ok := os.LookupEnv("HEARTBEAT_INTERVAL")
	if !ok {
		logger.Warn("heartbeat interval not found")
		logger.Warn("heartbeats will be disabled")
	}

	var heartbeatDuration time.Duration
	if heartbeatInterval != "" {
		parsedHeartbeat, err := time.ParseDuration(heartbeatInterval)
		if err != nil {
			logger.Error(err.Error())
			panic(err)
		}
		heartbeatDuration = parsedHeartbeat
	}

	return worker.WorkerConf{
		GRPCAddr:          grpcAddr,
		WorkerAddr:        workerAddr,
		SchedulerAddr:     schedulerAddr,
		Logger:            logger,
		HeartbeatInterval: heartbeatDuration,
	}
}

func defaultLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, nil))
}
