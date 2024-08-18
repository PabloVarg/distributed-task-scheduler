package scheduler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/pablovarg/distributed-task-scheduler/internal/task"
)

type CreateTaskRequest struct {
	Command     string     `json:"command"`
	ScheduledAt *time.Time `json:"scheduled_at"`
	Delay       *string    `json:"delay"`
}

func (r *CreateTaskRequest) Valid() (map[string]string, bool) {
	validationErrors := make(map[string]string)

	if r.Command == "" {
		validationErrors["command"] = "command is required"
	}

	if r.ScheduledAt == nil && r.Delay == nil {
		validationErrors["scheduled_time"] = "you must specify a schedule time"
	}

	if r.ScheduledAt != nil && r.Delay != nil {
		validationErrors["scheduled_time"] = "you must only specify one schedule time"
	}

	if r.Delay != nil && r.ScheduledAt == nil {
		_, err := time.ParseDuration(*r.Delay)
		if err != nil {
			validationErrors["delay"] = "could not parse delay"
		}
	}

	return validationErrors, len(validationErrors) == 0
}

func (s *Scheduler) createTask(w http.ResponseWriter, r *http.Request) {
	var request CreateTaskRequest

	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	if validationErrors, valid := request.Valid(); !valid {
		if err := json.NewEncoder(w).Encode(validationErrors); err != nil {
			http.Error(w, "", http.StatusInternalServerError)
		}
		return
	}

	scheduledTime := time.Now()
	if request.ScheduledAt != nil {
		scheduledTime = *request.ScheduledAt
	}
	if request.Delay != nil {
		requestDelay, _ := time.ParseDuration(*request.Delay)

		scheduledTime = time.Now().Add(requestDelay)
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if _, err := s.taskModel.CreateTask(ctx, task.Task{
		Command:     request.Command,
		ScheduledAt: scheduledTime,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
