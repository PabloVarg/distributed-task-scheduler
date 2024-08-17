package main

import (
	"context"
	"log"
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

func run(ctx context.Context, logger *log.Logger) {
	schedulerConf := readConf(logger)
	scheduler, err := scheduler.NewScheduler(schedulerConf, logger)
	if err != nil {
		logger.Fatalln(err)
	}

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	<-scheduler.Start(ctx)
}

func readConf(logger *log.Logger) scheduler.SchedulerConf {
	dsn, ok := os.LookupEnv("POSTGRES_DSN")
	if !ok {
		logger.Fatalf("env variable %q not set", "POSTGRES_DSN")
	}

	return scheduler.SchedulerConf{
		Addr:         ":8000",
		GRPCAddr:     ":9000",
		DB_DSN:       dsn,
		Logger:       logger,
		PollInterval: 1 * time.Second,
		BatchSize:    100,
	}
}

func defaultLogger() *log.Logger {
	return log.New(os.Stdout, "", log.LUTC|log.Lshortfile)
}
