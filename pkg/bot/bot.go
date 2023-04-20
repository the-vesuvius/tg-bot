package bot

import (
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
	"strconv"
	"tg_bot/logger"
	"tg_bot/pkg/dao"
	"tg_bot/pkg/errs"
	"tg_bot/pkg/models"
)

type Bot struct {
	key      string
	botApi   *tgbotapi.BotAPI
	usersDao dao.Users
	tasksDao dao.Tasks
}

func NewBot(key string, usersDao dao.Users, tasksDao dao.Tasks) (*Bot, error) {
	bot, err := tgbotapi.NewBotAPI(key)
	if err != nil {
		return nil, err
	}

	return &Bot{
		key:      key,
		botApi:   bot,
		usersDao: usersDao,
		tasksDao: tasksDao,
	}, nil
}

func (b *Bot) Run() {
	logger.Get().Info("Bot is running")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.botApi.GetUpdatesChan(u)
	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				err := b.HandleStartCmd(update)
				if err != nil {
					logger.Get().Error("HandleStartCmd failed", zap.Error(err))
				}
			case "add":
				err := b.HandleAddCmd(update)
				if err != nil {
					logger.Get().Error("HandleAddCmd failed", zap.Error(err))
				}
			case "done":
				err := b.HandleDoneCmd(update)
				if err != nil {
					logger.Get().Error("HandleDoneCmd failed", zap.Error(err))
				}
			}
		}
	}
}

func (b *Bot) HandleStartCmd(update tgbotapi.Update) error {
	inputTgUserId := update.Message.From.ID
	tgUserId := strconv.FormatInt(inputTgUserId, 10)

	_, err := b.ensureUserExists(tgUserId)
	if err != nil {
		sendErr := b.SendMessage(update.Message.Chat.ID, "Something went wrong, please try again later")
		if sendErr != nil {
			logger.Get().Error("Could not send message", zap.Error(sendErr))
		}
		return err
	}

	err = b.SendMessage(update.Message.Chat.ID, "Hello, I'm @read_that_bot!")
	if err != nil {
		logger.Get().Error("Could not send message", zap.Error(err))
		return err
	}

	return nil
}

func (b *Bot) HandleAddCmd(update tgbotapi.Update) error {
	inputTgUserId := update.Message.From.ID
	tgUserId := strconv.FormatInt(inputTgUserId, 10)

	user, err := b.ensureUserExists(tgUserId)
	if err != nil {
		sendErr := b.SendMessage(update.Message.Chat.ID, "Something went wrong, please try again later")
		if sendErr != nil {
			logger.Get().Error("Could not send message", zap.Error(sendErr))
		}
		return err
	}

	task := models.Task{
		UserId: user.Id,
		Url:    update.Message.CommandArguments(),
		Status: models.TaskStatusNew,
	}

	_, err = b.tasksDao.InsertTask(&task)
	if err != nil {
		logger.Get().Error("Could not insert task", zap.Error(err))
		sendErr := b.SendMessage(update.Message.Chat.ID, "Something went wrong, please try again later")
		if sendErr != nil {
			logger.Get().Error("Could not send message", zap.Error(sendErr))
		}

		return err
	}

	err = b.SendMessage(update.Message.Chat.ID, "Task added successfully")
	if err != nil {
		logger.Get().Error("Could not send message", zap.Error(err))
		return err
	}
	return nil
}

func (b *Bot) HandleDoneCmd(update tgbotapi.Update) error {
	inputTgUserId := update.Message.From.ID
	tgUserId := strconv.FormatInt(inputTgUserId, 10)

	user, err := b.ensureUserExists(tgUserId)
	if err != nil {
		sendErr := b.SendMessage(update.Message.Chat.ID, "Something went wrong, please try again later")
		if sendErr != nil {
			logger.Get().Error("Could not send message", zap.Error(sendErr))
		}
		return err
	}

	tasks, err := b.tasksDao.GetInProgressTasksByUserId(user.Id)
	if err != nil {
		logger.Get().Error("Could not get tasks", zap.Error(err))
		sendErr := b.SendMessage(update.Message.Chat.ID, "Something went wrong, please try again later")
		if sendErr != nil {
			logger.Get().Error("Could not send message", zap.Error(sendErr))
		}

		return err
	}

	if len(tasks) == 0 {
		err = b.SendMessage(update.Message.Chat.ID, "You don't have any tasks in progress")
		if err != nil {
			logger.Get().Error("Could not send message", zap.Error(err))
			return err
		}
		return nil
	}

	var taskIds []int64
	for _, task := range tasks {
		taskIds = append(taskIds, task.Id)
	}

	err = b.tasksDao.UpdateTasksStatus(taskIds, models.TaskStatusDone)
	if err != nil {
		logger.Get().Error("Could not update tasks", zap.Error(err))
		sendErr := b.SendMessage(update.Message.Chat.ID, "Something went wrong, please try again later")
		if sendErr != nil {
			logger.Get().Error("Could not send message", zap.Error(sendErr))
		}

		return err
	}

	tasks, err = b.tasksDao.GetUsersTasksByStatus(user.Id, models.TaskStatusNew)
	if err != nil {
		logger.Get().Error("Could not get tasks", zap.Error(err))
		sendErr := b.SendMessage(update.Message.Chat.ID, "Something went wrong, please try again later")
		if sendErr != nil {
			logger.Get().Error("Could not send message", zap.Error(sendErr))
		}

		return err
	}

	err = b.SendMessage(update.Message.Chat.ID, fmt.Sprintf("Tasks marked as done successfully. You got %d task(s) left in backlog", len(tasks)))
	if err != nil {
		logger.Get().Error("Could not send message", zap.Error(err))
		return err
	}

	return nil
}

func (b *Bot) SendMessage(chatId int64, text string) error {
	msg := tgbotapi.NewMessage(chatId, text)
	_, err := b.botApi.Send(msg)
	if err != nil {
		return err
	}

	return nil
}

func (b *Bot) ensureUserExists(tgUserExternalId string) (*models.User, error) {
	var user *models.User
	user, err := b.usersDao.GetUserByExternalId(tgUserExternalId)
	if err != nil {
		if errors.Is(err, &errs.ErrNotFound{}) {
			user, err = b.usersDao.InsertUser(&models.User{
				ExternalId: tgUserExternalId,
			})
			if err != nil {
				logger.Get().Error("Could not insert user", zap.Error(err))
				return nil, err
			}
		} else {
			logger.Get().Error("Could not get user by external id", zap.Error(err))
			return nil, err
		}
	}

	return user, nil
}
