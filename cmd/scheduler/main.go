package main

import (
	"log"
	"os"

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
	app.logger.Println("scheduler started")

	<-scheduler.Start()
}

func (app *app) readConf() scheduler.SchedulerConf {
	dsn, ok := os.LookupEnv("POSTGRES_DSN")
	if !ok {
		app.logger.Fatalf("env variable %q not set", "POSTGRES_DSN")
	}

	return scheduler.SchedulerConf{
		DB_DSN: dsn,
	}
}

func defaultLogger() *log.Logger {
	return log.New(os.Stdout, "", log.LUTC|log.Lshortfile)
}
