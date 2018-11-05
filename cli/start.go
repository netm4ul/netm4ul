package cmd

import (
	"fmt"
	"os"

	"github.com/netm4ul/netm4ul/core/api"
	"github.com/netm4ul/netm4ul/core/client"
	"github.com/netm4ul/netm4ul/core/communication"
	"github.com/netm4ul/netm4ul/core/server"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	serverLogPath = "server.log"
	clientLogPath = "client.log"
)

var (
	cliLogfile bool
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the requested service",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("\n\nTo few arguments !\n\n")
		cmd.Help()
		os.Exit(1)
	},
}

// startServerCmd represents the startServer command
var startServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the server",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {

		var err error
		// init session...
		// there is no chaining of persistent pre run ... so we are doing it manually...
		createSessionBase()
		if cliLogfile {
			setupLoggingToFile(serverLogPath)
		}

		cliSession.IsServer = isServer
		cliSession.Nodes = make(map[string]communication.Node, 0)

		if cliSession.Config.TLSParams.UseTLS {
			cliSession.Config.TLSParams.TLSConfig, err = cliSession.Config.TLSBuildServerConf()

			if err != nil {
				log.Error("Unable to load TLS configuration. Shutting down.")
				os.Exit(1)
			}
		}

	},
	Run: func(cmd *cobra.Command, args []string) {

		// TODO : not sure if we should use the CLI session or a new one ...
		// ss := session.NewSession(config.Config)
		// listen on all interface + Server port
		s := server.CreateServer(cliSession)
		go s.Listen()
		// TODO flag enable / disable api
		// TODO : not sure if we should use the CLI session or a new one ...
		// sa := session.NewSession(config.Config)
		a := api.NewAPI(cliSession, s)
		go a.Start()

		gracefulShutdown()

	},
}

// startClientCmd represents the startServer command
var startClientCmd = &cobra.Command{
	Use:   "client",
	Short: "Start the client",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {

		var err error
		// init session
		// there is no chaining of persistent pre run ... so we are doing it manually...
		createSessionBase()
		if cliLogfile {
			setupLoggingToFile(clientLogPath)
		}
		cliSession.IsClient = isClient

		if cliSession.Config.TLSParams.UseTLS {
			cliSession.Config.TLSParams.TLSConfig, err = cliSession.Config.TLSBuildClientConf()

			if err != nil {
				log.Error("Unable to load TLS configuration. Shutting down.")
				os.Exit(1)
			}
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		// TODO : not sure if we should use the CLI session or a new one ...
		// sc := session.NewSession(config.Config)
		c := client.CreateClient(cliSession)
		go c.Start()

		gracefulShutdown()
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.PersistentFlags().BoolVar(&cliLogfile, "log2file", false, "Enable logging to file")

	startCmd.AddCommand(startServerCmd)
	startCmd.AddCommand(startClientCmd)
}
