package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"tg_bot/logger"
	"tg_bot/pkg/bot"
)

func InitRunCommand() *cobra.Command {
	var apiKey string

	var runCmd = &cobra.Command{
		Use:   "run",
		Short: "Run the bot",
		Run: func(cmd *cobra.Command, args []string) {
			apiKey, _ = os.LookupEnv("TG_BOT_API_KEY")
			if apiKey == "" {
				logger.Get().Error("Please set the TG_BOT_API_KEY environment variable")
				os.Exit(1)
			}
			botApp, err := bot.NewBot(apiKey)
			if err != nil {
				logger.Get().Error("Bot app could not be created", zap.Error(err))
				os.Exit(1)
			}

			var exit = make(chan os.Signal, 1)

			go func() {
				botApp.Run()
			}()

			signal.Notify(exit, os.Interrupt)

			<-exit

			logger.Get().Sync()
		},
	}

	return runCmd
}
