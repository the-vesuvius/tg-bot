package models

import "time"

type Task struct {
	Id        int64
	UserId    int
	Url       string
	CreatedAt time.Time
	UpdatedAt time.Time
}
