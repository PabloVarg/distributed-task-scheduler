package scheduler

import (
	"net/http"
)

func (s *Scheduler) routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /tasks", s.createTask)
	mux.HandleFunc("GET /tasks/{id}", s.retrieveTask)

	return mux
}
