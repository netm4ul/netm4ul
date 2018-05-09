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
	name       string
	shortName  string
	moduleType string
	author     string
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

		if name == "" {
			fmt.Println("You must provide an adapter name")
			cmd.Help()
			os.Exit(1)
		}
		generate.GenerateAdapter(name, shortName)
	},
}

var createModuleCmd = &cobra.Command{
	Use:   "module",
	Short: "Generate a new module",
	Run: func(cmd *cobra.Command, args []string) {
		allowedType := map[string]bool{
			"recon":   true,
			"report":  true,
			"exploit": true,
		}

		if name == "" {
			fmt.Println("You must provide a module name")
			cmd.Help()
			os.Exit(1)
		}

		if _, ok := allowedType[moduleType]; ok {
			generate.Module(name, shortName, moduleType, author)
		} else {
			fmt.Println("Unknown type of module")
			cmd.Help()
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.AddCommand(createAdapterCmd)
	createCmd.AddCommand(createModuleCmd)
	createCmd.PersistentFlags().StringVar(&name, "name", "", "Name used for the folder and struct")
	createCmd.PersistentFlags().StringVar(&shortName, "short-name", "", "Short name used for the instancied struct")
	createCmd.PersistentFlags().StringVar(&author, "author", "", "Author name")
	createCmd.PersistentFlags().StringVar(&moduleType, "type", "", "Type of the new module")
}
