package errs

import "tg_bot/pkg/models"

type ErrNotFinished struct {
	Task *models.Task
}

func NewErrNotFinished(task *models.Task) *ErrNotFinished {
	return &ErrNotFinished{Task: task}
}

func (e *ErrNotFinished) Error() string {
	return "Task not finished, Task: " + e.Task.Url
}

func (e *ErrNotFinished) Is(target error) bool {
	_, ok := target.(*ErrNotFinished)
	return ok
}
