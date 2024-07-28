package scheduler

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pablovarg/distributed-task-scheduler/internal/task"
)

type SchedulerConf struct {
	DB_DSN string
	Logger *log.Logger
	Addr   string
}

type Scheduler struct {
	db        *sqlx.DB
	logger    *log.Logger
	addr      string
	taskModel task.TaskModel
}

func NewScheduler(conf SchedulerConf, logger *log.Logger) (*Scheduler, error) {
	db, err := sqlx.Open("postgres", conf.DB_DSN)
	if err != nil {
		return nil, fmt.Errorf("can not connect to the database (%w)", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("can not connect to the database (%w)", err)
	}

	assignedLogger := conf.Logger
	if assignedLogger == nil {
		assignedLogger = log.New(io.Discard, "", 0)
	}

	return &Scheduler{
		db:     db,
		logger: assignedLogger,
		addr:   conf.Addr,
		taskModel: task.TaskModel{
			DB: db,
		},
	}, nil
}

func (s *Scheduler) Start() error {
	return s.startServer()
}

func (s *Scheduler) startServer() error {
	srv := http.Server{
		Addr:         s.addr,
		Handler:      s.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	s.logger.Printf("Listening on %s", s.addr)
	return srv.ListenAndServe()
}
