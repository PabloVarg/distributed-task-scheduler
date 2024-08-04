package scheduler

import (
	"context"
	"fmt"
	"io"
	"log"
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
	Logger       *log.Logger
	Addr         string
	GRPCAddr     string
	PollInterval time.Duration
	BatchSize    int
}

type Scheduler struct {
	db        *sqlx.DB
	logger    *log.Logger
	taskModel task.TaskModel
	SchedulerConf
}

func NewScheduler(conf SchedulerConf, logger *log.Logger) (*Scheduler, error) {
	db, err := sqlx.Connect("postgres", conf.DB_DSN)
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
		taskModel: task.TaskModel{
			DB: db,
		},
		SchedulerConf: conf,
	}, nil
}

func (s *Scheduler) Start(ctx context.Context) <-chan any {
	done := make(chan any)

	go s.pollTasks(ctx)
	go s.startServer(ctx)
	go s.startGRPCServer(ctx)

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
		s.logger.Printf("Listening on %s\n", s.SchedulerConf.Addr)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			s.logger.Fatalln(err)
		}
		s.logger.Println("Shutting down server")
	}()

	select {
	case <-ctx.Done():
		srv.Shutdown(ctx)
	}
}

func (s *Scheduler) pollTasks(ctx context.Context) {
	ticker := time.NewTicker(s.SchedulerConf.PollInterval)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			tasks, err := s.taskModel.GetDueTasks(ctx, s.SchedulerConf.BatchSize)
			if err != nil {
				s.logger.Fatalln(err)
			}

			for _, task := range tasks {
				err := s.taskModel.CompleteTask(ctx, task.ID)
				if err != nil {
					s.logger.Println(err)
				}
			}
		}
	}
}

func (s *Scheduler) startGRPCServer(ctx context.Context) {
	lis, err := net.Listen("tcp", s.SchedulerConf.GRPCAddr)
	if err != nil {
		s.logger.Fatalln(err)
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

	s.logger.Printf("Grpc server listening on %s", s.SchedulerConf.GRPCAddr)
	if err := server.Serve(lis); err != nil {
		s.logger.Fatalln(err)
	}
}
