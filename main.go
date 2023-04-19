package main

import (
	"github.com/spf13/cobra"
	"tg_bot/cmd"
)

func main() {
	var rootCmd = &cobra.Command{
		Use: "app",
	}

	rootCmd.AddCommand(
		cmd.InitRunCommand(),
	)

	err := rootCmd.Execute()
	if err != nil {
		panic(err)
	}
}
