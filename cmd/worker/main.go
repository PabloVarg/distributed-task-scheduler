package main

import (
	"context"
	"log"
	"os"

	"github.com/pablovarg/distributed-task-scheduler/worker"
)

type app struct {
	logger *log.Logger
}

func main() {
	logger := log.New(os.Stdout, "", log.LUTC|log.Lshortfile)
	app := app{
		logger: logger,
	}
	conf := app.readConf(logger)
	worker := worker.NewWorker(conf)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app.logger.Println("Starting worker")
	<-worker.Start(ctx)
	app.logger.Println("Shutting down worker")
}

func (app *app) readConf(logger *log.Logger) worker.WorkerConf {
	schedulerAddr, ok := os.LookupEnv("SCHEDULER_ADDR")
	if !ok {
		app.logger.Fatalln("scheduler address not found")
	}
	workerAddr, ok := os.LookupEnv("WORKER_ADDR")
	if !ok {
		app.logger.Fatalln("worker address not found")
	}

	app.logger.Printf("scheduler addr: %s identified", schedulerAddr)
	return worker.WorkerConf{
		WorkerAddr:    workerAddr,
		SchedulerAddr: schedulerAddr,
		Logger:        logger,
	}
}
