package scheduler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/pablovarg/distributed-task-scheduler/internal/task"
)

func (s *Scheduler) createTask(w http.ResponseWriter, r *http.Request) {
	var request CreateTaskRequest

	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	if valid := DecodeValidator(w, request); !valid {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	task, err := s.taskModel.CreateTask(ctx, task.Task{
		Command:     request.Command,
		ScheduledAt: futureTime(request.ScheduledAt, request.Delay),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.logger.Info("created task", "task", task)
	w.WriteHeader(http.StatusCreated)
}
