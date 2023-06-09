package models

import "time"

const (
	TaskStatusNew        = "NEW"
	TaskStatusInProgress = "IN_PROGRESS"
	TaskStatusDone       = "DONE"
)

type TaskStatus string

func (s TaskStatus) String() string {
	return string(s)
}

type Task struct {
	Id        int64
	UserId    int64
	Url       string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}
