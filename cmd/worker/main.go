package main

import (
	"context"
	"log"
	"os"

	"github.com/pablovarg/distributed-task-scheduler/worker"
)

func main() {
	run(defaultLogger())
}

func run(logger *log.Logger) {
	ctx, cancel := context.WithCancel(context.Background())
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
