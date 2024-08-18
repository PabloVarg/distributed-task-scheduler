package scheduler

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"

	pb "github.com/pablovarg/distributed-task-scheduler/internal/grpc"
	"github.com/pablovarg/distributed-task-scheduler/internal/task"
)

type SchedulerConf struct {
	DB_DSN       string
	Logger       *slog.Logger
	Addr         string
	GRPCAddr     string
	PollInterval time.Duration
	BatchSize    int
}

type Scheduler struct {
	db        *sqlx.DB
	logger    *slog.Logger
	taskModel task.TaskModel
	WorkerPool
	SchedulerConf
}

func NewScheduler(conf SchedulerConf) (*Scheduler, error) {
	db, err := sqlx.Connect("postgres", conf.DB_DSN)
	if err != nil {
		return nil, fmt.Errorf("can not connect to the database [%w]", err)
	}

	assignedLogger := conf.Logger
	if assignedLogger == nil {
		assignedLogger = slog.New(slog.NewJSONHandler(io.Discard, nil))
	}

	return &Scheduler{
		db:     db,
		logger: assignedLogger,
		taskModel: task.TaskModel{
			DB: db,
		},
		WorkerPool: WorkerPool{
			workers: make(map[string]*Worker),
			logger:  assignedLogger,
		},
		SchedulerConf: conf,
	}, nil
}

func (s *Scheduler) Start(ctx context.Context) <-chan any {
	done := make(chan any)

	go s.pollTasks(ctx)
	go s.startServer(ctx)
	go s.startGRPCServer(ctx)
	go s.cleanWorkersContext(ctx)

	go func(ctx context.Context) {
		select {
		case <-ctx.Done():
			close(done)
		}
	}(ctx)

	return done
}

func (s *Scheduler) startServer(ctx context.Context) {
	srv := http.Server{
		Addr:         s.SchedulerConf.Addr,
		Handler:      s.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	go func() {
		s.logger.Info(fmt.Sprintf("listening on %s", s.SchedulerConf.Addr))
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			s.logger.Error(err.Error())
			panic(err)
		}
		s.logger.Info("shutting down server")
	}()

	select {
	case <-ctx.Done():
		srv.Shutdown(ctx)
	}
}

func (s *Scheduler) pollTasks(ctx context.Context) {
	ticker := time.NewTicker(s.SchedulerConf.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(ctx, s.SchedulerConf.PollInterval)
			defer cancel()

			tasks, err := s.taskModel.GetDueTasks(ctx, s.SchedulerConf.BatchSize)
			if err != nil {
				s.logger.Error(err.Error())
				panic(err)
			}

			for _, task := range tasks {
				s.logger.Info("processing task", "task", task)
				workerId, err := s.nextWorker()
				if err != nil {
					s.logger.Error("could not retrieve next worker")
					continue
				}

				s.WorkerPool.RLock()
				worker, ok := s.workers[workerId]
				if !ok {
					s.logger.Error("could not retrieve selected worker")
					continue
				}

				go s.sendTask(task, worker)

				s.WorkerPool.RUnlock()
			}
		}
	}
}

func (s *Scheduler) sendTask(task task.Task, worker *Worker) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := worker.client.ExecuteJob(ctx, &pb.Task{
		ID:      int64(task.ID),
		Command: task.Command,
	})
	if err != nil {
		s.logger.Error(err.Error())
	}
}

func (s *Scheduler) startGRPCServer(ctx context.Context) {
	lis, err := net.Listen("tcp", s.SchedulerConf.GRPCAddr)
	if err != nil {
		s.logger.Error(err.Error())
	}

	server := grpc.NewServer()
	pb.RegisterSchedulerServer(server, &SchedulerServerImpl{
		Scheduler: s,
	})

	go func(ctx context.Context) {
		select {
		case <-ctx.Done():
			server.GracefulStop()
		}
	}(ctx)

	s.logger.Info(fmt.Sprintf("grpc server listening on %s", s.SchedulerConf.GRPCAddr))
	if err := server.Serve(lis); err != nil {
		s.logger.Error(err.Error())
		panic(err)
	}
}
