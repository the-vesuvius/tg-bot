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
