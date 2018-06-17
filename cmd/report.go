package cmd

import (
	"fmt"

	"github.com/netm4ul/netm4ul/modules/report"
	"github.com/spf13/cobra"
)

var (
	reportType string
)

// reportCmd represents the report command
var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate a new report",
	PreRun: func(cmd *cobra.Command, args []string) {
		report.LoadReports()
	},
	Run: func(cmd *cobra.Command, args []string) {
		err := report.Reporter["text"].Generate("test")
		if err != nil {
			fmt.Printf("Could not generated this report : %s\n", err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(reportCmd)
	reportCmd.PersistentFlags().StringVar(&name, "name", "", "Name of the report [default : <project>-<date>.<extension>]")
	reportCmd.PersistentFlags().StringVar(&reportType, "type", "", "Type of report [text, pdf...]")
}
