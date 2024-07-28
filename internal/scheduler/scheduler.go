package scheduler

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type SchedulerConf struct {
	DB_DSN string
	Logger *log.Logger
}

type Scheduler struct {
	db     *sqlx.DB
	logger *log.Logger
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

	return &Scheduler{
		db:     db,
		logger: conf.Logger,
	}, nil
}

func (s *Scheduler) Start() error {
	return s.startServer()
}

func (s *Scheduler) startServer() error {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /tasks", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "create task")
	})

	srv := http.Server{
		Addr:         ":8000",
		Handler:      mux,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	s.logger.Println("Listening on port 8080")
	return srv.ListenAndServe()
}
