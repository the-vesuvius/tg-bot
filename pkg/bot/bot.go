package bot

import (
	"errors"
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
}

func NewBot(key string, usersDao dao.Users) (*Bot, error) {
	bot, err := tgbotapi.NewBotAPI(key)
	if err != nil {
		return nil, err
	}

	return &Bot{
		key:      key,
		botApi:   bot,
		usersDao: usersDao,
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
			}
		}
	}
}

func (b *Bot) HandleStartCmd(update tgbotapi.Update) error {
	inputTgUserId := update.Message.From.ID
	tgUserId := strconv.FormatInt(inputTgUserId, 10)

	_, err := b.usersDao.GetUserByExternalId(tgUserId)
	if err != nil {
		if errors.Is(err, &errs.ErrNotFound{}) {
			_, err = b.usersDao.InsertUser(&models.User{
				ExternalId: tgUserId,
			})
			if err != nil {
				logger.Get().Error("Could not insert user", zap.Error(err))
				return err
			}
		} else {
			logger.Get().Error("Could not get user by external id", zap.Error(err))
			return err
		}
	}

	err = b.SendMessage(update.Message.Chat.ID, "Hello, I'm @read_that_bot!")
	if err != nil {
		logger.Get().Error("Could not send message", zap.Error(err))
		return err
	}

	return nil
}

func (b *Bot) HandleAddCmd(update tgbotapi.Update) error {
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
