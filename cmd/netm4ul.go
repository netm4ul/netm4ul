package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/netm4ul/netm4ul/core/config"
	"github.com/spf13/cobra"
)

const (
	DefaultConfigPath = "netm4ul.conf"
)

var (
	Modes       = []string{"passive", "stealth", "aggressive"}
	DefaultMode = Modes[1] // uses first non-passive mode.

	configPath     string
	CLItargets     []string
	CLImodules     []string
	CLImode        string
	CLIProjectName string
	verbose        bool
	version        bool

	isServer   bool
	isClient   bool
	noColors   bool
	info       string
	completion bool
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", DefaultConfigPath, "Custom config file path")
	rootCmd.PersistentFlags().StringVarP(&CLIProjectName, "project", "p", DefaultConfigPath, "Uses the provided project name")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVarP(&noColors, "no-colors", "", false, "Disable color printing")
}

var rootCmd = &cobra.Command{
	Use:   "netm4ul",
	Short: "netm4ul : Distributed recon made easy",

	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		config.LoadConfig(configPath)

		config.Config.ConfigPath = configPath
		config.Config.Verbose = verbose
		config.Config.Project.Name = CLIProjectName

	},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		os.Exit(1)
	},
}

// Execute is the entrypoint
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func gracefulShutdown() {
	// handle graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	log.Println("shutting down")
	os.Exit(0)
}
