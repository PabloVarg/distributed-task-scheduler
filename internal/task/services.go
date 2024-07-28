package task

import (
	"context"
)

func (m *TaskModel) CreateTask(ctx context.Context, task Task) (*Task, error) {
	query := "INSERT INTO task (command) VALUES ($1) RETURNING id, created_at"

	if err := m.DB.QueryRowContext(ctx, query, task.Command).Scan(&task.ID, &task.CreatedAt); err != nil {
		return nil, err
	}

	return &task, nil
}
