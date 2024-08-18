package scheduler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/pablovarg/distributed-task-scheduler/internal/api"
	"github.com/pablovarg/distributed-task-scheduler/internal/task"
)

func (s *Scheduler) createTask(w http.ResponseWriter, r *http.Request) {
	var request CreateTaskRequest

	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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
	w.Header().Set("Location", strconv.Itoa(task.ID))
	w.WriteHeader(http.StatusCreated)
}

func (s *Scheduler) retrieveTask(w http.ResponseWriter, r *http.Request) {
	id, err := api.IntPathValue(r, "id")
	if err != nil || id < 1 {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	task, err := s.taskModel.GetTask(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			http.NotFound(w, r)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if err := json.NewEncoder(w).Encode(task); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
