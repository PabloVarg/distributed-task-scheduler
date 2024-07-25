package scheduler

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type SchedulerConf struct {
	DB_DSN string
	logger *log.Logger
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
		logger: conf.logger,
	}, nil
}

func (s *Scheduler) Start() <-chan any {
	done := make(chan any)
	close(done)
	return done
}
