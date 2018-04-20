// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"

	"github.com/netm4ul/netm4ul/core/api"
	"github.com/netm4ul/netm4ul/core/client"
	"github.com/netm4ul/netm4ul/core/config"
	"github.com/netm4ul/netm4ul/core/server"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the requested service",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("\n\nTo few arguments !\n\n")
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
		// there is no chaining of persistent pre run ... so we are doing it manualy...
		createSessionBase()

		CLISession.Config.IsServer = isServer
		CLISession.Config.Nodes = make(map[string]config.Node)

		if CLISession.Config.TLSParams.UseTLS {
			CLISession.Config.TLSParams.TLSConfig, err = config.TLSBuildServerConf()

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
		go server.CreateServer(CLISession)

		// TODO flag enable / disable api
		// TODO : not sure if we should use the CLI session or a new one ...
		// sa := session.NewSession(config.Config)
		go api.CreateAPI(CLISession)

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
		// there is no chaining of persistent pre run ... so we are doing it manualy...
		createSessionBase()

		config.Config.IsClient = isClient

		if CLISession.Config.TLSParams.UseTLS {
			config.Config.TLSParams.TLSConfig, err = config.TLSBuildClientConf()

			if err != nil {
				log.Error("Unable to load TLS configuration. Shutting down.")
				os.Exit(1)
			}
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		// TODO : not sure if we should use the CLI session or a new one ...
		// sc := session.NewSession(config.Config)
		go client.CreateClient(CLISession)

		gracefulShutdown()
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.AddCommand(startServerCmd)
	startCmd.AddCommand(startClientCmd)
}
