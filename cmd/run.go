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
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/netm4ul/netm4ul/modules"

	log "github.com/sirupsen/logrus"
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
			log.Errorf("Error while parsing targets : %v", err.Error())
		}

		log.Debugf("targets : %+v", targets)
		log.Debugf("CLIModules : %+v", CLImodules)
		log.Debugf("Modules : %+v", CLISession.Config.Modules)
		log.Debugf("CLIMode : %+v", CLImode)
		log.Debugf("Mode : %+v", CLISession.Config.Mode)

		if len(CLImodules) > 0 {
			mods, err := parseModules(CLImodules, CLISession)
			if err != nil {
				log.Errorf(err.Error())
			}
			addModules(mods, CLISession)
		}
		runModules(targets)
	},
}

func runModules(targets []modules.Input) {
	url := "http://" + CLISession.Config.Server.IP + ":" + strconv.FormatUint(uint64(CLISession.Config.API.Port), 10) + "/api/v1/projects/FirstProject/run/"

	jsonInput, err := json.Marshal(targets)
	if err != nil {
		log.Fatal(err)
	}

	for i := range CLISession.Config.Modules {
		resp, err := http.Post(url+i, "application/text", bytes.NewBuffer(jsonInput))
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
