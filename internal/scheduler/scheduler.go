package scheduler

type SchedulerConf struct {
	DB_DSN string
}

type Scheduler struct {
	conf SchedulerConf
}

func NewScheduler(conf SchedulerConf) *Scheduler {
	return &Scheduler{
		conf: conf,
	}
}

func (s *Scheduler) Start() <-chan any {
	done := make(chan any)

	close(done)

	return done
}
