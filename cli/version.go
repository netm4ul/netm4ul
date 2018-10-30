package cmd

import (
	"github.com/netm4ul/netm4ul/cli/ui"
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints version",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		createSessionBase()
	},
	Run: func(cmd *cobra.Command, args []string) {
		ui.PrintVersion(cliSession)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
