package main

import (
	"log"
	"os"
)

type app struct {
	logger *log.Logger
}

func main() {
	logger := log.New(os.Stdout, "", log.LUTC|log.Lshortfile)
	app := app{
		logger: logger,
	}

	app.readConf()
}

func (app *app) readConf() {
	schedulerAddr, ok := os.LookupEnv("SCHEDULER_ADDR")
	if !ok {
		app.logger.Fatalln("scheduler address not found")
	}

	app.logger.Printf("scheduler addr: %s", schedulerAddr)
}
