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
	//DefaultConfigPath represent the default config file. It may be replaced by the user
	// TOFIX : We should check multiple path with a fixed priority (./netm4ul.conf, /etc/netm4ul/netm4ul.conf, ~/.config/netm4ul/netm4ul.conf for example)
	DefaultConfigPath = "netm4ul.conf"
)

var (
	//Modes represents the different levels of informations netm4ul will attempts
	// - passive doens'nt do any action on the server
	// - stealth do some information gathering directly on the host and slow down the rate
	// - aggressive perform all the scans and uses an high requests rate
	Modes = []string{"passive", "stealth", "aggressive"}

	// DefaultMode pick mode in case of missing config informations
	DefaultMode = Modes[1] // uses first non-passive mode.

	configPath     string
	cliTargets     []string
	cliModules     []string
	cliMode        string
	cliProjectName string
	verbose        bool
	version        bool

	isServer   bool
	isClient   bool
	noColors   bool
	info       string
	completion bool
	cliSession *session.Session
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", DefaultConfigPath, "Custom config file path")
	rootCmd.PersistentFlags().StringVarP(&cliProjectName, "project", "p", "", "Uses the provided project name")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVarP(&noColors, "no-colors", "", false, "Disable color printing")
	log.SetOutput(os.Stdout)
	customFormatter := new(log.TextFormatter)
	// customFormatter.TimestampFormat = "2001-02-03 12:34:56"
	customFormatter.FullTimestamp = true
	log.SetFormatter(customFormatter)
}

func createSessionBase() {
	if verbose {
		log.SetLevel(log.DebugLevel)
	}

	cfg, err := config.LoadConfig(configPath)

	if err != nil {
		log.Fatalf("Could not load the config file : %s. Please provide a config file (-config path/to/configfile). Error : %s", configPath, err)
	}
	// this function will fill the missing value if the config file is missing field
	// will be especially useful during updates
	setDefaultValues(&cfg)

	cliSession, err = session.NewSession(cfg)
	if err != nil {
		log.Fatalf("Could not create the CLI session : %s", err)
	}
	cliSession.ConfigPath = configPath
	cliSession.Verbose = verbose

	if cliProjectName != "" {
		cliSession.Config.Project.Name = cliProjectName
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
