package cmd

import (
	"fmt"
	"os"

	"github.com/netm4ul/netm4ul/cli/ui"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Return all results",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("To few arguments !")
		cmd.Help()
		os.Exit(1)
	},
}

var listProjectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "Return list of projects",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		createSessionBase()
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug("listProjectsCmd called")
		ui.PrintProjectsInfo(cliSession)
	},
}

var listProjectCmd = &cobra.Command{
	Use:   "project",
	Short: "Return project info",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		createSessionBase()
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug("listProjectCmd called")
		// no argument, read from config
		if len(args) == 0 {
			ui.PrintProjectInfo(cliSession.Config.Project.Name, cliSession)
			os.Exit(0)
		}
		// 1 arguments, use it
		if len(args) == 1 {
			ui.PrintProjectInfo(args[0], cliSession)
		} else {
			fmt.Println("Too many arguments expected 1, got", len(args))
		}
	},
}

var listIPCmd = &cobra.Command{
	Use:   "ip",
	Short: "Return ip info",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		createSessionBase()
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug("listIPCmd called")
	},
}

var listPortCmd = &cobra.Command{
	Use:   "port",
	Short: "Return port info",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		createSessionBase()
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug("listPortCmd called")
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.AddCommand(listProjectsCmd)
	listCmd.AddCommand(listProjectCmd)
	listCmd.AddCommand(listIPCmd)
	listCmd.AddCommand(listPortCmd)
}
