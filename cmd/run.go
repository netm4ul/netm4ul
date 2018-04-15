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
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/netm4ul/netm4ul/cmd/colors"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run scan on the defined target",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		createSessionBase()
		CLISession.Config.Mode = CLImode
	},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			log.Fatalln("To few arguments ! Expecting target.")
		}

		targets, err := parseTargets(args)
		if err != nil {
			log.Println("Error while parsing targets :", err.Error())
		}

		if CLISession.Config.Verbose {
			log.Println("targets :", targets)
			log.Println("CLIModules :", CLImodules)
			log.Println("Modules :", CLISession.Config.Modules)
			log.Println("CLIMode :", CLImode)
			log.Println("Mode :", CLISession.Config.Mode)
		}

		if len(CLImodules) > 0 {
			mods, err := parseModules(CLImodules, CLISession)
			if err != nil {
				fmt.Println(colors.Yellow(err.Error()))
			}
			addModules(mods, CLISession)
		}
		runModules(args[0])
	},
}

func runModules(target string) {
	url := "http://" + CLISession.Config.Server.IP + ":" + strconv.FormatUint(uint64(CLISession.Config.API.Port), 10) + "/api/v1/projects/FirstProject/run/"
	for i := range CLISession.Config.Modules {
		resp, err := http.Post(url+i+"?options="+target, "application/text", strings.NewReader("127.0.0.1"))
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(resp)
	}
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.PersistentFlags().StringArrayVar(&CLImodules, "modules", []string{}, "Set custom module(s)")
	runCmd.PersistentFlags().StringVarP(&CLImode, "mode", "m", DefaultMode, "Use predefined mode")

}
