package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/pablovarg/distributed-task-scheduler/worker"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	run(ctx, defaultLogger())
}

func run(ctx context.Context, logger *log.Logger) {
	ctx, cancel := signal.NotifyContext(ctx)
	defer cancel()

	conf := readConf(logger)
	<-worker.NewWorker(conf).Start(ctx)
}

func readConf(logger *log.Logger) worker.WorkerConf {
	schedulerAddr, ok := os.LookupEnv("SCHEDULER_ADDR")
	if !ok {
		logger.Fatalln("scheduler address not found")
	}
	workerAddr, ok := os.LookupEnv("WORKER_ADDR")
	if !ok {
		logger.Fatalln("worker address not found")
	}

	return worker.WorkerConf{
		GRPCAddr:      ":9000",
		WorkerAddr:    workerAddr,
		SchedulerAddr: schedulerAddr,
		Logger:        logger,
	}
}

func defaultLogger() *log.Logger {
	return log.New(os.Stdout, "", log.LUTC|log.Lshortfile)
}
