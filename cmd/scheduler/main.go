package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
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

	return scheduler.SchedulerConf{
		Addr:             ":8000",
		GRPCAddr:         ":9000",
		DB_DSN:           dsn,
		Logger:           logger,
		PollInterval:     1 * time.Second,
		BatchSize:        100,
		WorkerDeadPeriod: parsedWorkerDeadPeriod,
	}
}

func defaultLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, nil))
}
