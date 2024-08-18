package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"time"

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
	dsn, ok := os.LookupEnv("POSTGRES_DSN")
	if !ok {
		err := fmt.Errorf("env variable %s not set", "POSTGRES_DSN")

		logger.Error(err.Error())
		panic(err)
	}
	apiAddr, ok := os.LookupEnv("API_ADDR")
	if !ok {
		err := fmt.Errorf("env variable %s not set", "API_ADDR")

		logger.Error(err.Error())
		panic(err)
	}
	grpcAddr, ok := os.LookupEnv("GRPC_ADDR")
	if !ok {
		err := fmt.Errorf("env variable %s not set", "GRPC_ADDR")

		logger.Error(err.Error())
		panic(err)
	}

	var parsedBatchSize int
	batchSize, ok := os.LookupEnv("BATCH_SIZE")
	if ok {
		value, err := strconv.Atoi(batchSize)
		if err != nil {
			err := fmt.Errorf("could not parse BATCH_SIZE")

			logger.Error(err.Error())
			panic(err)
		}

		parsedBatchSize = value
	}

	workerPeriod, ok := os.LookupEnv("WORKER_DEAD_PERIOD")
	if !ok {
		err := fmt.Errorf("env variable %s not set", "WORKER_DEAD_PERIOD")

		logger.Error(err.Error())
		panic(err)
	}
	parsedWorkerDeadPeriod, err := time.ParseDuration(workerPeriod)
	if err != nil {
		err := fmt.Errorf("could not parse WORKER_DEAD_PERIOD")

		logger.Error(err.Error())
		panic(err)
	}

	pollInterval, ok := os.LookupEnv("POLL_INTERVAL")
	if !ok {
		err := fmt.Errorf("env variable %s not set", "POLL_INTERVAL")

		logger.Error(err.Error())
		panic(err)
	}
	parsedPollInterval, err := time.ParseDuration(pollInterval)
	if err != nil {
		err := fmt.Errorf("could not parse POLL_INTERVAL")

		logger.Error(err.Error())
		panic(err)
	}

	return scheduler.SchedulerConf{
		Addr:             apiAddr,
		GRPCAddr:         grpcAddr,
		DB_DSN:           dsn,
		Logger:           logger,
		PollInterval:     parsedPollInterval,
		BatchSize:        parsedBatchSize,
		WorkerDeadPeriod: parsedWorkerDeadPeriod,
	}
}

func defaultLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, nil))
}
