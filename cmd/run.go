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

	"gopkg.in/mgo.v2"

	"github.com/netm4ul/netm4ul/core/api"
	"github.com/netm4ul/netm4ul/core/database"
	"github.com/netm4ul/netm4ul/modules"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var projectName string

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run scan on the defined target",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		createSessionBase()
		CLISession.Config.Mode = CLImode
		var mgoSession *mgo.Session
		mgoSession = database.Connect()
		database.CreateProject(mgoSession, projectName, "projects")
	},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			log.Fatalln("Too few arguments ! Expecting target.")
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

		// if len(CLImodules) > 0 {
		// 	mods, err := parseModules(CLImodules, CLISession)
		// 	if err != nil {
		// 		log.Errorf(err.Error())
		// 	}
		// 	addModules(mods, CLISession)
		// }

		if len(CLImodules) > 0 {
			fmt.Println("Running only specified module(s) :", CLImodules)
			runSpecifiedModules(targets, CLImodules)
			return
		}

		runModules(targets)
	},
}

func runModules(targets []modules.Input) {
	url := "http://" + CLISession.Config.Server.IP + ":" + strconv.FormatUint(uint64(CLISession.Config.API.Port), 10) + "/api/v1/projects/FirstProject/run"

	jsonInput, err := json.Marshal(targets)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonInput))
	if err != nil {
		log.Fatal(err)
	}
	res, err := json.Marshal(resp.Body)
	if err != nil {
		fmt.Println("Recieved invalid json !", err)
	}
	fmt.Println(res)
}

func runSpecifiedModules(targets []modules.Input, modules []string) {

	url := "http://" + CLISession.Config.Server.IP + ":" + strconv.FormatUint(uint64(CLISession.Config.API.Port), 10) + "/api/v1/projects/FirstProject/run"

	jsonInput, err := json.Marshal(targets)
	if err != nil {
		log.Fatal(err)
	}

	for _, m := range modules {
		r, err := http.Post(url+"/"+m, "application/json", bytes.NewBuffer(jsonInput))
		if err != nil {
			log.Fatal(err)

		}
		var res api.Result
		err = json.NewDecoder(r.Body).Decode(&res)
		if err != nil {
			fmt.Println("Recieved invalid json !", err)
		}
		if res.Code == api.CodeOK {
			fmt.Println(res.Message, res.Data)
		} else {
			fmt.Printf("An error occured : [%d], %s\n", res.Code, res.Message)
		}
	}
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.PersistentFlags().StringArrayVar(&CLImodules, "modules", []string{}, "Set custom module(s)")
	runCmd.PersistentFlags().StringVarP(&CLImode, "mode", "m", DefaultMode, "Use predefined mode")
	runCmd.PersistentFlags().StringVarP(&projectName, "project", "", "default_project", "Set project name")
}
