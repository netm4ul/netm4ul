package cmd

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/netm4ul/netm4ul/core/config"
	"github.com/netm4ul/netm4ul/core/session"
	log "github.com/sirupsen/logrus"
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
	CLIprojectName string
	verbose        bool
	version        bool

	isServer   bool
	isClient   bool
	noColors   bool
	info       string
	completion bool
	CLISession *session.Session
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", DefaultConfigPath, "Custom config file path")
	rootCmd.PersistentFlags().StringVarP(&CLIprojectName, "project", "p", "", "Uses the provided project name")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVarP(&noColors, "no-colors", "", false, "Disable color printing")
	log.SetOutput(os.Stdout)
	customFormatter := new(log.TextFormatter)
	// customFormatter.TimestampFormat = "2001-02-03 12:34:56"
	customFormatter.FullTimestamp = true
	log.SetFormatter(customFormatter)
}

func createSessionBase() {
	var cfg config.ConfigToml
	if verbose {
		log.SetLevel(log.DebugLevel)
	}

	err := config.LoadConfig(configPath)
	cfg = config.Config

	if err != nil {
		log.Debugf("Could not load the config file : %s", configPath)
		cfg = config.ConfigToml{}
	}

	CLISession = session.NewSession(cfg)
	CLISession.ConfigPath = configPath
	CLISession.Verbose = verbose

	if CLIprojectName != "" {
		CLISession.Config.Project.Name = CLIprojectName
	}
}

var rootCmd = &cobra.Command{
	Use:   "netm4ul",
	Short: "netm4ul : Distributed recon made easy",

	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		createSessionBase()
	},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		os.Exit(1)
	},
}

// Execute is the entrypoint
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

func gracefulShutdown() {
	// handle graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	log.Info("shutting down")
	os.Exit(0)
}

func setupLoggingToFile(logfile string) {
	if logfile != "" {
		f, err := os.OpenFile(logfile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to initialize log file %s", err)
			os.Exit(1)
		}
		log.SetOutput(f)
	}
}
