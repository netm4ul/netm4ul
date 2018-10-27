package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of the requested service",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("To few arguments !")
		cmd.Help()
		os.Exit(1)
	},
}

// statusServerCmd represents the statusServer command
var statusServerCmd = &cobra.Command{
	Use:   "server",
	Short: "status of the server",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		createSessionBase()
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("statusServer called")
	},
}

// statusClientCmd represents the statusServer command
var statusClientCmd = &cobra.Command{
	Use:   "client",
	Short: "status of the client",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		createSessionBase()
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("statusClient called")
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
	statusCmd.AddCommand(statusServerCmd)
	statusCmd.AddCommand(statusClientCmd)
}
