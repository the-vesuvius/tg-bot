package bot

import (
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
	"strconv"
	"strings"
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
			case "current":
				err := b.HandleCurrentCmd(update)
				if err != nil {
					logger.Get().Error("HandleCurrentCmd failed", zap.Error(err))
				}
			case "next":
				err := b.HandleNextCmd(update)
				if err != nil {
					logger.Get().Error("HandleNextCmd failed", zap.Error(err))
				}
			case "skip":
				err := b.HandleSkipCmd(update)
				if err != nil {
					logger.Get().Error("HandleSkipCmd failed", zap.Error(err))
				}
			}
		}
	}
}

func (b *Bot) HandleStartCmd(update tgbotapi.Update) error {
	inputTgUserId := update.Message.From.ID
	tgUserId := strconv.FormatInt(inputTgUserId, 10)

	_, err := b.ensureUserExists(tgUserId, update.Message.Chat.ID)
	if err != nil {
		sendErr := b.SendMessage(update.Message.Chat.ID, "Something went wrong, please try again later")
		if sendErr != nil {
			logger.Get().Error("Could not send message", zap.Error(sendErr))
		}
		return err
	}

	err = b.SendMessage(update.Message.Chat.ID, "Hello, I'm @read_that_bot!\n"+
		"I will remind you to read your articles from your reading list(at 5pm UTC, time currently isn't configurable).\n"+
		"Use /add <article url> command to add new article to your reading list.\n"+
		"Use /current command to get current article from your reading list.\n"+
		"Use /done command to mark current article as read.\n"+
		"Use /next command to get next article from your reading list(if you don't want to wait for the next time I remind you).\n",
	)
	if err != nil {
		logger.Get().Error("Could not send message", zap.Error(err))
		return err
	}

	return nil
}

func (b *Bot) HandleAddCmd(update tgbotapi.Update) error {
	inputTgUserId := update.Message.From.ID
	tgUserId := strconv.FormatInt(inputTgUserId, 10)

	user, err := b.ensureUserExists(tgUserId, update.Message.Chat.ID)
	if err != nil {
		sendErr := b.SendMessage(update.Message.Chat.ID, "Something went wrong, please try again later")
		if sendErr != nil {
			logger.Get().Error("Could not send message", zap.Error(sendErr))
		}
		return err
	}

	taskUrl := update.Message.CommandArguments()
	taskUrl = strings.Trim(taskUrl, " ")
	if taskUrl == "" {
		sendErr := b.SendMessage(update.Message.Chat.ID, "Please provide article url")
		if sendErr != nil {
			logger.Get().Error("Could not send message", zap.Error(sendErr))
		}
		return errors.New("empty task url")
	}

	task := models.Task{
		UserId: user.Id,
		Url:    taskUrl,
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

	user, err := b.ensureUserExists(tgUserId, update.Message.Chat.ID)
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

func (b *Bot) HandleCurrentCmd(update tgbotapi.Update) error {
	inputTgUserId := update.Message.From.ID
	tgUserId := strconv.FormatInt(inputTgUserId, 10)

	user, err := b.ensureUserExists(tgUserId, update.Message.Chat.ID)
	if err != nil {
		sendErr := b.SendMessage(update.Message.Chat.ID, "Something went wrong, please try again later")
		if sendErr != nil {
			logger.Get().Error("Could not send message", zap.Error(sendErr))
		}
		return err
	}

	tasks, err := b.tasksDao.GetUsersTasksByStatus(user.Id, models.TaskStatusInProgress)
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

	task := tasks[0]
	err = b.SendMessage(update.Message.Chat.ID, fmt.Sprintf("Your current task is %s", task.Url))
	if err != nil {
		logger.Get().Error("Could not send message", zap.Error(err))
		return err
	}

	return nil
}

func (b *Bot) HandleNextCmd(update tgbotapi.Update) error {
	inputTgUserId := update.Message.From.ID
	tgUserId := strconv.FormatInt(inputTgUserId, 10)

	user, err := b.ensureUserExists(tgUserId, update.Message.Chat.ID)
	if err != nil {
		sendErr := b.SendMessage(update.Message.Chat.ID, "Something went wrong, please try again later")
		if sendErr != nil {
			logger.Get().Error("Could not send message", zap.Error(sendErr))
		}
		return err
	}

	task, err := b.GetNextTask(user)
	if err != nil {
		if errors.Is(err, &errs.ErrNotFinished{}) {
			var notFinishedErr *errs.ErrNotFinished
			errors.As(err, &notFinishedErr)
			sendErr := b.SendMessage(update.Message.Chat.ID, "You have unfinished task. Please finish it first. Your current task is \n"+notFinishedErr.Task.Url)
			if sendErr != nil {
				logger.Get().Error("Could not send message", zap.Error(sendErr))
				return err
			}
			return sendErr
		}

		if errors.Is(err, &errs.ErrNotFound{}) {
			sendErr := b.SendMessage(update.Message.Chat.ID, "There is no tasks available. Please add some tasks first")
			if sendErr != nil {
				logger.Get().Error("Could not send message", zap.Error(sendErr))
				return sendErr
			}
			return sendErr
		}

		sendErr := b.SendMessage(update.Message.Chat.ID, "Something went wrong, please try again later")
		if sendErr != nil {
			logger.Get().Error("Could not send message", zap.Error(sendErr))
		}

		return err
	}

	err = b.SendMessage(update.Message.Chat.ID, fmt.Sprintf("Your next task is: \n%s", task.Url))
	if err != nil {
		logger.Get().Error("Could not send message", zap.Error(err))
		return err
	}

	return nil
}

func (b *Bot) HandleSkipCmd(update tgbotapi.Update) error {
	inputTgUserId := update.Message.From.ID
	tgUserId := strconv.FormatInt(inputTgUserId, 10)

	user, err := b.ensureUserExists(tgUserId, update.Message.Chat.ID)
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

	if len(tasks) >= 0 {
		var taskIds []int64
		for _, task := range tasks {
			taskIds = append(taskIds, task.Id)
		}

		err = b.tasksDao.UpdateTasksStatus(taskIds, models.TaskStatusNew)
		if err != nil {
			logger.Get().Error("Could not update tasks", zap.Error(err))
			sendErr := b.SendMessage(update.Message.Chat.ID, "Something went wrong, please try again later")
			if sendErr != nil {
				logger.Get().Error("Could not send message", zap.Error(sendErr))
			}

			return err
		}
	}

	task, err := b.GetNextTask(user)
	if err != nil {
		if errors.Is(err, &errs.ErrNotFinished{}) {
			var notFinishedErr *errs.ErrNotFinished
			errors.As(err, &notFinishedErr)
			sendErr := b.SendMessage(update.Message.Chat.ID, "You have unfinished task. Please finish it first. Your current task is \n"+notFinishedErr.Task.Url)
			if sendErr != nil {
				logger.Get().Error("Could not send message", zap.Error(sendErr))
				return err
			}
			return sendErr
		}

		if errors.Is(err, &errs.ErrNotFound{}) {
			sendErr := b.SendMessage(update.Message.Chat.ID, "There is no tasks available. Please add some tasks first")
			if sendErr != nil {
				logger.Get().Error("Could not send message", zap.Error(sendErr))
				return sendErr
			}
			return sendErr
		}

		sendErr := b.SendMessage(update.Message.Chat.ID, "Something went wrong, please try again later")
		if sendErr != nil {
			logger.Get().Error("Could not send message", zap.Error(sendErr))
		}

		return err
	}

	err = b.SendMessage(update.Message.Chat.ID, fmt.Sprintf("Your next task is: \n%s", task.Url))
	if err != nil {
		logger.Get().Error("Could not send message", zap.Error(err))
		return err
	}

	return nil
}

func (b *Bot) GetNextTask(user *models.User) (*models.Task, error) {
	inProgressTasks, err := b.tasksDao.GetUsersTasksByStatus(user.Id, models.TaskStatusInProgress)
	if err != nil {
		logger.Get().Error("Could not get tasks", zap.Error(err))
		return nil, err
	}

	if len(inProgressTasks) > 0 {
		return nil, errs.NewErrNotFinished(inProgressTasks[0])
	}

	task, err := b.tasksDao.GetUsersRandomTaskByStatus(user.Id, models.TaskStatusNew)
	if err != nil {
		if errors.Is(err, &errs.ErrNotFound{}) {
			return nil, err
		}
		logger.Get().Error("Could not get random task", zap.Error(err))
		return nil, err
	}

	err = b.tasksDao.UpdateTasksStatus([]int64{task.Id}, models.TaskStatusInProgress)
	if err != nil {
		logger.Get().Error("Could not update task status", zap.Error(err))
		return nil, err
	}

	return task, nil
}

func (b *Bot) SendReminders() error {
	users, err := b.usersDao.GetAllUsers()
	if err != nil {
		logger.Get().Error("Could not get users", zap.Error(err))
		return err
	}

	for _, user := range users {
		task, err := b.GetNextTask(user)
		if err != nil {
			if errors.Is(err, &errs.ErrNotFinished{}) {
				var notFinishedErr *errs.ErrNotFinished
				errors.As(err, &notFinishedErr)
				err = b.SendMessage(user.ChatId, "You have unfinished task. Please finish it first. Your current task is \n"+notFinishedErr.Task.Url)
				if err != nil {
					logger.Get().Error("Could not send message", zap.Error(err))
				}
				continue
			}

			if errors.Is(err, &errs.ErrNotFound{}) {
				continue
			}

			err = b.SendMessage(user.ChatId, "Something went wrong, please try again later")
			if err != nil {
				logger.Get().Error("Could not send message", zap.Error(err))
			}
			continue
		}

		err = b.SendMessage(user.ChatId, fmt.Sprintf("Your next task is: \n%s", task.Url))
		if err != nil {
			logger.Get().Error("Could not send message", zap.Error(err))
			continue
		}
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

func (b *Bot) ensureUserExists(tgUserExternalId string, chatId int64) (*models.User, error) {
	var user *models.User
	user, err := b.usersDao.GetUserByExternalId(tgUserExternalId)
	if err != nil {
		if errors.Is(err, &errs.ErrNotFound{}) {
			user, err = b.usersDao.InsertUser(&models.User{
				ExternalId: tgUserExternalId,
				ChatId:     chatId,
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
