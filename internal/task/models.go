package task

import (
	"time"

	"github.com/jmoiron/sqlx"
)

type Task struct {
	ID           int        `json:"id" db:"id"`
	Command      string     `json:"command" db:"command"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	ScheduledAt  time.Time  `json:"scheduled_at" db:"scheduled_at"`
	SuccessfulAt *time.Time `json:"successful_at" db:"successful_at"`
}

type TaskModel struct {
	DB *sqlx.DB
}
