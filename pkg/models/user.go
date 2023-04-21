package models

import "time"

type User struct {
	Id         int64
	ExternalId string
	ChatId     int64
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
