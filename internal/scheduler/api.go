package scheduler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/pablovarg/distributed-task-scheduler/internal/task"
)

func (s *Scheduler) createTask(w http.ResponseWriter, r *http.Request) {
	request := struct {
		Command     string    `json:"command"`
		ScheduledAt time.Time `json:"scheduled_at"`
	}{}

	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	if request.Command == "" {
		http.Error(w, "{\"message\": \"a command should be given\"}", http.StatusUnprocessableEntity)
	}

	if request.ScheduledAt.Before(time.Now()) {
		http.Error(w, "{\"message\": \"scheduled time must be in the future\"}", http.StatusUnprocessableEntity)
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	s.taskModel.CreateTask(ctx, task.Task{
		Command:     request.Command,
		ScheduledAt: request.ScheduledAt,
	})
	w.WriteHeader(http.StatusCreated)
}
