package main

import (
	"fmt"
	"os"

	"harvest-cli/cmd"

	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "h",
		Short: "Harvest CLI is a time tracking utility",
		Long: `A simple CLI utility to log time entries for your work.
Complete documentation is available at https://github.com/amanangira/harvest-cli-utility`,
		Run: func(cmd *cobra.Command, args []string) {
			// If no subcommand is provided, print help
			cmd.Help()
		},
	}

	// Add commands
	rootCmd.AddCommand(cmd.CreateCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
