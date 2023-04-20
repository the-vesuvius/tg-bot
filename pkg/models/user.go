package models

import "time"

type User struct {
	Id         int64
	ExternalId string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
