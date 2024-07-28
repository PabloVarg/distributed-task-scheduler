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
		Command string `json:"command"`
	}{}

	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	if request.Command == "" {
		http.Error(w, "{\"message\": \"a command should be given\"}", http.StatusUnprocessableEntity)
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	s.taskModel.CreateTask(ctx, task.Task{
		Command: request.Command,
	})
	w.WriteHeader(http.StatusCreated)
}
