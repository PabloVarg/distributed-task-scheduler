package task

import (
	"context"
	"fmt"
)

func (m *TaskModel) CreateTask(ctx context.Context, task Task) (*Task, error) {
	query := `
        INSERT INTO
            task (command, scheduled_at)
        VALUES
            ($1, $2)
        RETURNING
            id, created_at
    `

	var result Task
	if err := m.DB.GetContext(ctx, &result, query, task.Command, task.ScheduledAt); err != nil {
		return nil, err
	}

	return &result, nil
}

func (m *TaskModel) GetDueTasks(ctx context.Context, batchSize int) ([]Task, error) {
	query := `
        SELECT
            id, command, created_at, scheduled_at, successful_at
        FROM
            task
        WHERE
            scheduled_at <= NOW() + INTERVAL '15 days'
            AND successful_at IS NULL
            AND picked_at IS NULL
            AND failed_at IS NULL
    `

	if batchSize != 0 {
		query += `
            LIMIT $1
        `
	}

	query += `
        FOR UPDATE SKIP LOCKED
    `

	var dueTasks []Task
	switch batchSize {
	case 0:
		if err := m.DB.SelectContext(ctx, &dueTasks, query); err != nil {
			return nil, err
		}
	default:
		if err := m.DB.SelectContext(ctx, &dueTasks, query, batchSize); err != nil {
			return nil, err
		}
	}

	return dueTasks, nil
}

func (m *TaskModel) PickTask(ctx context.Context, taskID int) error {
	return m.updateTask(ctx, taskID, "picked_at")
}

func (m *TaskModel) CompleteTask(ctx context.Context, taskID int) error {
	return m.updateTask(ctx, taskID, "successful_at")
}

func (m *TaskModel) FailTask(ctx context.Context, taskID int) error {
	return m.updateTask(ctx, taskID, "failed_at")
}

func (m *TaskModel) updateTask(ctx context.Context, taskID int, field string) error {
	selectQuery := `
        SELECT id FROM task WHERE id = $1
    `

	updateQuery := fmt.Sprintf(`
        UPDATE
            task
        SET
            %s = NOW()
        WHERE
            id = $1
    `, field)

	tx, err := m.DB.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, selectQuery, taskID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("task not found")
	}

	tx.ExecContext(ctx, updateQuery, taskID)

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
