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

	"github.com/netm4ul/netm4ul/cmd/colors"
	"github.com/netm4ul/netm4ul/core/config"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run scan on the defined target",
	PreRun: func(cmd *cobra.Command, args []string) {
		config.Config.Mode = CLImode
	},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			log.Fatalln("To few arguments ! Expecting target.")
		}

		targets, err := parseTargets(args)
		if err != nil {
			log.Println("Error while parsing targets :", err.Error())
		}

		if config.Config.Verbose {
			log.Println("targets :", targets)
			log.Println("CLIModules :", CLImodules)
			log.Println("Modules :", config.Config.Modules)
			log.Println("CLIMode :", CLImode)
			log.Println("Mode :", config.Config.Mode)
		}

		if len(CLImodules) > 0 {
			mods, err := parseModules(CLImodules)
			if err != nil {
				fmt.Println(colors.Yellow(err.Error()))
			}
			addModules(mods)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.PersistentFlags().StringArrayVar(&CLImodules, "modules", []string{}, "Set custom module(s)")
	runCmd.PersistentFlags().StringVarP(&CLImode, "mode", "m", DefaultMode, "Use predefined mode")

}
