package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

func InitRunCommand() *cobra.Command {
	var runCmd = &cobra.Command{
		Use:   "run",
		Short: "Run the bot",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("run called")
		},
	}

	return runCmd
}
