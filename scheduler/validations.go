package scheduler

import (
	"encoding/json"
	"net/http"
	"time"
)

type Validator interface {
	Valid() (map[string]string, bool)
}

func DecodeValidator(w http.ResponseWriter, validator Validator) bool {
	validationErrors, valid := validator.Valid()
	if valid {
		return valid
	}

	if err := json.NewEncoder(w).Encode(validationErrors); err != nil {
		http.Error(w, "", http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusUnprocessableEntity)

	return valid
}

type CreateTaskRequest struct {
	Command     string     `json:"command"`
	ScheduledAt *time.Time `json:"scheduled_at"`
	Delay       *string    `json:"delay"`
}

func (r CreateTaskRequest) Valid() (map[string]string, bool) {
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
