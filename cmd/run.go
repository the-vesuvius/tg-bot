package cmd

import (
	"database/sql"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"tg_bot/db"
	"tg_bot/logger"
	"tg_bot/pkg/bot"
	"tg_bot/pkg/dao"
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
			dbUser, _ := os.LookupEnv("TG_BOT_DB_USER")
			if dbUser == "" {
				logger.Get().Error("Please set the TG_BOT_DB_USER environment variable")
				os.Exit(1)
			}
			dbPass, _ := os.LookupEnv("TG_BOT_DB_PASSWORD")
			if dbPass == "" {
				logger.Get().Error("Please set the TG_BOT_DB_PASSWORD environment variable")
				os.Exit(1)
			}
			dbHost, _ := os.LookupEnv("TG_BOT_DB_HOST")
			if dbHost == "" {
				logger.Get().Error("Please set the TG_BOT_DB_HOST environment variable")
				os.Exit(1)
			}
			dbPort, _ := os.LookupEnv("TG_BOT_DB_PORT")
			if dbPort == "" {
				logger.Get().Error("Please set the TG_BOT_DB_PORT environment variable")
				os.Exit(1)
			}

			dbConnUrl := fmt.Sprintf("%s:%s@tcp(%s:%s)/read_that_bot?parseTime=true", dbUser, dbPass, dbHost, dbPort)
			dbConn, err := sql.Open("mysql", dbConnUrl)
			if err != nil {
				logger.Get().Error("DB connection failed", zap.Error(err))
				os.Exit(1)
			}
			defer dbConn.Close()

			usersDao := dao.NewUsers(dbConn)

			botApp, err := bot.NewBot(apiKey, usersDao)
			if err != nil {
				logger.Get().Error("Bot app could not be created", zap.Error(err))
				os.Exit(1)
			}

			migrator := db.NewMigrator(dbConnUrl)
			err = migrator.Migrate()
			if err != nil {
				if err == migrate.ErrNoChange {
					logger.Get().Info("DB already migrated")
				} else {
					logger.Get().Error("DB migration failed", zap.Error(err))
					os.Exit(1)
				}
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
