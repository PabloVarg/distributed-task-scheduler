package task

import (
	"time"

	"github.com/jmoiron/sqlx"
)

type Task struct {
	ID           int        `json:"id" db:"id"`
	Command      string     `json:"command" db:"command"`
	ScheduledAt  time.Time  `json:"scheduled_at" db:"scheduled_at"`
	PickedAt     *time.Time `db:"picked_at"`
	SuccessfulAt *time.Time `json:"" db:"successful_at"`
	FailedAt     *time.Time `db:"failed_at"`
	CreatedAt    time.Time  `json:"" db:"created_at"`
}

type TaskModel struct {
	DB *sqlx.DB
}
