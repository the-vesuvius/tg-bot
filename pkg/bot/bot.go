package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
	"tg_bot/logger"
)

type Bot struct {
	key    string
	botApi *tgbotapi.BotAPI
}

func NewBot(key string) (*Bot, error) {
	bot, err := tgbotapi.NewBotAPI(key)
	if err != nil {
		return nil, err
	}

	return &Bot{
		key:    key,
		botApi: bot,
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

		logger.Get().Info("Received message", zap.Any("MSG", update.Message))
	}
}

func (b *Bot) HandleStartCmd(update tgbotapi.Update) error {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Hello, I'm @read_that_bot!")
	msg.ReplyToMessageID = update.Message.MessageID
	_, err := b.botApi.Send(msg)
	if err != nil {
		return err
	}

	return nil
}

func (b *Bot) HandleAddCmd(update tgbotapi.Update) error {
	return nil
}
