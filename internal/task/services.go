package task

import (
	"context"
	"fmt"
	"reflect"
)

func (m *TaskModel) GetTask(ctx context.Context, ID int) (*Task, error) {
	query := `
        SELECT
            id, command, scheduled_at, picked_at, successful_at, failed_at, created_at
        FROM
            task
        WHERE
            id = $1
    `

	var result Task
	if err := m.DB.GetContext(ctx, &result, query, ID); err != nil {
		return nil, err
	}

	return &result, nil
}

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
            scheduled_at <= NOW() + INTERVAL '1 seconds'
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
	return m.updateTask(ctx, taskID, "picked_at", "PickedAt")
}

func (m *TaskModel) CompleteTask(ctx context.Context, taskID int) error {
	return m.updateTask(ctx, taskID, "successful_at", "SuccessfulAt")
}

func (m *TaskModel) FailTask(ctx context.Context, taskID int) error {
	return m.updateTask(ctx, taskID, "failed_at", "FailedAt")
}

func (m *TaskModel) updateTask(ctx context.Context, taskID int, dbField string, structField string) error {
	selectQuery := `
        SELECT id, picked_at, successful_at, failed_at FROM task WHERE id = $1
    `

	updateQuery := fmt.Sprintf(`
        UPDATE
            task
        SET
            %s = NOW()
        WHERE
            id = $1
    `, dbField)

	tx, err := m.DB.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var task Task
	err = tx.GetContext(ctx, &task, selectQuery, taskID)
	if err != nil {
		return err
	}

	if !reflect.ValueOf(task).FieldByName(structField).IsNil() {
		return fmt.Errorf("state %s already set", structField)
	}

	tx.ExecContext(ctx, updateQuery, taskID)

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
