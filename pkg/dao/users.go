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

type Users interface {
	InsertUser(user *models.User) (*models.User, error)
	GetUserById(userId int64) (*models.User, error)
	GetUserByExternalId(externalId string) (*models.User, error)
}

type users struct {
	db *sql.DB
}

func NewUsers(db *sql.DB) *users {
	return &users{db: db}
}

func (u *users) InsertUser(user *models.User) (*models.User, error) {
	query := sq.Insert("users").Columns("external_id", "chat_id").
		Values(user.ExternalId, user.ChatId)

	res, err := query.RunWith(u.db).Exec()
	if err != nil {
		return nil, err
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	newUser, err := u.GetUserById(lastId)
	if err != nil {
		logger.Get().Error("Could not get user by id", zap.Error(err))
		return nil, err
	}

	return newUser, nil
}

func (u *users) GetUserById(userId int64) (*models.User, error) {
	query := sq.Select("id", "external_id", "created_at", "updated_at").
		From("users").
		Where(sq.Eq{"id": userId})

	rows, err := query.RunWith(u.db).Query()
	if err != nil {
		return nil, err
	}

	var user models.User
	if rows.Next() {
		err = rows.Scan(&user.Id, &user.ExternalId, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
	} else {
		idStr := strconv.FormatInt(userId, 10)

		return nil, errs.NewErrNotFound("User", "id", idStr)
	}

	return &user, nil
}

func (u *users) GetUserByExternalId(externalId string) (*models.User, error) {
	query := sq.Select("id", "external_id", "created_at", "updated_at").
		From("users").
		Where(sq.Eq{"external_id": externalId})

	rows, err := query.RunWith(u.db).Query()
	if err != nil {
		return nil, err
	}

	var user models.User
	if rows.Next() {
		err = rows.Scan(&user.Id, &user.ExternalId, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errs.NewErrNotFound("User", "external_id", externalId)
	}

	return &user, nil
}
