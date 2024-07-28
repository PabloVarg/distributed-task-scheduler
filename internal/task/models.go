package task

import (
	"time"

	"github.com/jmoiron/sqlx"
)

type Task struct {
	ID           int       `json:"id"`
	Command      string    `json:"command"`
	CreatedAt    time.Time `json:"created_at"`
	ScheduledAt  time.Time `json:"scheduled_at"`
	SuccessfulAt time.Time `json:"successful_at"`
}

type TaskModel struct {
	DB *sqlx.DB
}
