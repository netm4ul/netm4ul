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

	"github.com/netm4ul/netm4ul/scripts/generate"
	"github.com/spf13/cobra"
)

var (
	adapterName      string
	adapterShortName string
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create the requested ressource",

	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		createSessionBase()
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("create called")
	},
}

var createAdapterCmd = &cobra.Command{
	Use:   "adapter",
	Short: "Generate a new adapter",
	Run: func(cmd *cobra.Command, args []string) {

		if adapterName == "" {
			fmt.Println("You must provide an adapter name")
			cmd.Help()
			os.Exit(1)
		}
		generate.GenerateAdapter(adapterName, adapterShortName)
	},
}

func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.AddCommand(createAdapterCmd)
	createAdapterCmd.Flags().StringVar(&adapterName, "name", "", "Adapter name (folder and struct)")
	createAdapterCmd.Flags().StringVar(&adapterShortName, "short-name", "", "Adapter short name (name of struct)")
}
