package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/pablovarg/distributed-task-scheduler/scheduler"
)

func main() {
	run(defaultLogger())
}

func run(logger *log.Logger) {
	schedulerConf := readConf(logger)
	scheduler, err := scheduler.NewScheduler(schedulerConf, logger)
	if err != nil {
		logger.Fatalln(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
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
