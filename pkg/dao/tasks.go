package dao

import (
	"database/sql"
	sq "github.com/Masterminds/squirrel"
	"go.uber.org/zap"
	"strconv"
	"tg_bot/logger"
	"tg_bot/pkg/errs"
	"tg_bot/pkg/models"
)

type Tasks interface {
	InsertTask(task *models.Task) (*models.Task, error)
	GetTaskById(taskId int64) (*models.Task, error)
	GetInProgressTasksByUserId(userId int64) ([]*models.Task, error)
	UpdateTasksStatus(taskIds []int64, status string) error
	GetUsersTasksByStatus(userId int64, status string) ([]*models.Task, error)
}

type tasks struct {
	db *sql.DB
}

func NewTasks(db *sql.DB) *tasks {
	return &tasks{db: db}
}

func (t *tasks) InsertTask(task *models.Task) (*models.Task, error) {
	query := sq.Insert("tasks").Columns("user_id", "url", "status").
		Values(task.UserId, task.Url, task.Status)

	res, err := query.RunWith(t.db).Exec()
	if err != nil {
		return nil, err
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	newTask, err := t.GetTaskById(lastId)
	if err != nil {
		logger.Get().Error("Could not get task by id", zap.Error(err))
		return nil, err
	}
	return newTask, nil
}

func (t *tasks) GetTaskById(taskId int64) (*models.Task, error) {
	query := sq.Select("id", "user_id", "url", "status", "created_at", "updated_at").
		From("tasks").
		Where(sq.Eq{"id": taskId})

	rows, err := query.RunWith(t.db).Query()
	if err != nil {
		return nil, err
	}

	var task models.Task
	if rows.Next() {
		err := rows.Scan(&task.Id, &task.UserId, &task.Url, &task.Status, &task.CreatedAt, &task.UpdatedAt)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errs.NewErrNotFound("Task", "id", strconv.FormatInt(taskId, 10))
	}

	return &task, nil
}

func (t *tasks) GetInProgressTasksByUserId(userId int64) ([]*models.Task, error) {
	query := sq.Select("id", "user_id", "url", "status", "created_at", "updated_at").
		From("tasks").
		Where(sq.Eq{"user_id": userId}).
		Where(sq.Eq{"status": models.TaskStatusInProgress})

	rows, err := query.RunWith(t.db).Query()
	if err != nil {
		return nil, err
	}

	var tasksList = make([]*models.Task, 0)
	for rows.Next() {
		var task models.Task
		err := rows.Scan(&task.Id, &task.UserId, &task.Url, &task.Status, &task.CreatedAt, &task.UpdatedAt)
		if err != nil {
			return nil, err
		}
		tasksList = append(tasksList, &task)
	}

	return tasksList, nil
}

func (t *tasks) UpdateTasksStatus(taskIds []int64, status string) error {
	query := sq.Update("tasks").
		Set("status", status).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": taskIds})

	_, err := query.RunWith(t.db).Exec()
	if err != nil {
		return err
	}

	return nil
}

func (t *tasks) GetUsersTasksByStatus(userId int64, status string) ([]*models.Task, error) {
	query := sq.Select("id", "user_id", "url", "status", "created_at", "updated_at").
		From("tasks").
		Where(sq.Eq{"user_id": userId}).
		Where(sq.Eq{"status": status})

	rows, err := query.RunWith(t.db).Query()
	if err != nil {
		return nil, err
	}

	var tasksList = make([]*models.Task, 0)
	for rows.Next() {
		var task models.Task
		err := rows.Scan(&task.Id, &task.UserId, &task.Url, &task.Status, &task.CreatedAt, &task.UpdatedAt)
		if err != nil {
			return nil, err
		}
		tasksList = append(tasksList, &task)
	}

	return tasksList, nil
}
