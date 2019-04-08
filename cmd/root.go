package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "octopus",
	Short: "Octopus is a kubectl plugin that helps you do almost anything :D",
}

// Execute is the main execute for command
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Printf(`Error in executing octopus, error=%s`, err.Error())
		os.Exit(1)
	}
}
