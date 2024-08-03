package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/pablovarg/distributed-task-scheduler/internal/scheduler"
)

type app struct {
	logger *log.Logger
}

func main() {
	app := app{
		logger: defaultLogger(),
	}
	schedulerConf := app.readConf()

	scheduler, err := scheduler.NewScheduler(schedulerConf, app.logger)
	if err != nil {
		app.logger.Fatalln(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app.logger.Println("Starting scheduler")
	<-scheduler.Start(ctx)
	app.logger.Println("Shutting down scheduler")
}

func (app *app) readConf() scheduler.SchedulerConf {
	dsn, ok := os.LookupEnv("POSTGRES_DSN")
	if !ok {
		app.logger.Fatalf("env variable %q not set", "POSTGRES_DSN")
	}

	return scheduler.SchedulerConf{
		Addr:         ":8000",
		DB_DSN:       dsn,
		Logger:       app.logger,
		PollInterval: 1 * time.Second,
		// BatchSize:    10,
	}
}

func defaultLogger() *log.Logger {
	return log.New(os.Stdout, "", log.LUTC|log.Lshortfile)
}
