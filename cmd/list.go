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

	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Return all results",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("To few arguments !")
		cmd.Help()
		os.Exit(1)
	},
}

var listProjectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "Return list of projects",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		createSessionBase()
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("listProjectsCmd called")
		printProjectsInfo(CLISession)
	},
}

var listProjectCmd = &cobra.Command{
	Use:   "project",
	Short: "Return project info",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		createSessionBase()
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("listProjectCmd called")
		// no argument, read from config
		if len(args) == 0 {
			printProjectInfo(CLISession.Config.Project.Name, CLISession)
			os.Exit(0)
		}
		// 1 arguments, use it
		if len(args) == 1 {
			printProjectInfo(args[0], CLISession)
		} else {
			fmt.Println("Too many arguments expected 1, got", len(args))
		}

	},
}

var listIPCmd = &cobra.Command{
	Use:   "ip",
	Short: "Return ip info",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		createSessionBase()
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("listIPCmd called")
	},
}

var listPortCmd = &cobra.Command{
	Use:   "port",
	Short: "Return port info",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		createSessionBase()
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("listPortCmd called")
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.AddCommand(listProjectsCmd)
	listCmd.AddCommand(listProjectCmd)
	listCmd.AddCommand(listIPCmd)
	listCmd.AddCommand(listPortCmd)
}
