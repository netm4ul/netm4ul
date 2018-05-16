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
	"io"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/BurntSushi/toml"
	"github.com/netm4ul/netm4ul/core/database"
	"github.com/spf13/cobra"
)

type PromptRes struct {
	Message      string
	DefaultValue string
}

const (
	defaultDBSetupUser     = "admin"
	defaultDBSetupPassword = "admin"
	dbname                 = "NetM4ul"
)

var (
	cliDBSetupUser     string
	cliDBSetupPassword string
	userIMode          = false
	passwordIMode      = false
)

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "NetM4ul setup",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		createSessionBase()
		CLISession.Config.Database.User = cliDBSetupUser
		CLISession.Config.Database.Password = cliDBSetupPassword
	},

	Run: func(cmd *cobra.Command, args []string) {
		// ex2Conf()
		if CLISession.Config.Database.User == "" {
			userIMode = true
		}
		if CLISession.Config.Database.Password == "" {
			passwordIMode = true
		}
		setupDB()
	},
}

func check(err error) {
	if err != nil {
		log.Error(err)
	}
}

// prompt user for configuration parameters
func prompt(param string) (answer string) {
	// var text string
	var input string

	// Database parameters
	promptString := map[string]PromptRes{
		"dbuser":     {Message: "Interactive mode for database username (default : %s) : ", DefaultValue: defaultDBSetupUser},
		"dbpassword": {Message: "Interactive mode for database password (default : %s) : ", DefaultValue: defaultDBSetupPassword},
	}

	//Other parameters
	/*
		...
	*/

	fmt.Printf(promptString[param].Message, promptString[param].DefaultValue)
	fmt.Scanln(&input)
	if input == "" {
		return promptString[param].DefaultValue
	}

	return input
}

// setupDB => create user in db for future requests
func setupDB() {
	// TODO : check if the db isn't already setup (and use previous creds)
	CLISession.Config.Database.User = ""
	CLISession.Config.Database.Password = ""

	db := database.NewDatabase(&CLISession.Config)

	if userIMode {
		CLISession.Config.Database.User = prompt("dbuser")
	}

	if passwordIMode {
		CLISession.Config.Database.Password = prompt("dbpassword")
	}

	username := CLISession.Config.Database.User
	password := CLISession.Config.Database.Password

	// db.Connect(&CLISession.Config)
	err := db.SetupAuth(username, password, dbname)

	check(err)

	modifyDBConnect()
}

// modify conf file for db parameters
func modifyDBConnect() {
	//Open current config file
	srcFile, err := os.Open(CLISession.Config.ConfigPath)
	check(err)
	defer srcFile.Close()

	//Create bkp file
	destFile, err := os.Create(CLISession.Config.ConfigPath + ".old") // creates if file doesn't exist
	check(err)
	defer destFile.Close()

	//Copy content of current into bkp
	_, err = io.Copy(destFile, srcFile) // check first var for number of bytes copied
	check(err)

	err = destFile.Sync()
	check(err)

	//Remove old config file
	err = os.Remove(CLISession.Config.ConfigPath)
	check(err)

	//Create new config file
	file, err := os.OpenFile(
		CLISession.Config.ConfigPath,
		os.O_WRONLY|os.O_TRUNC|os.O_CREATE,
		0666,
	)
	check(err)
	defer file.Close()

	//Write new config to new config file
	if err := toml.NewEncoder(file).Encode(CLISession.Config); err != nil {
		log.Fatal(err)
	}
}

// modify conf file
func ex2Conf() {
	//Open example config file
	srcFile, err := os.Open(CLISession.Config.ConfigPath + ".example")
	check(err)
	defer srcFile.Close()

	//Open current config file
	destFile, err := os.OpenFile(
		CLISession.Config.ConfigPath,
		os.O_WRONLY|os.O_TRUNC|os.O_CREATE,
		0666,
	)
	check(err)
	defer destFile.Close()

	//Copy content of example into current
	_, err = io.Copy(destFile, srcFile) // check first var for number of bytes copied
	check(err)

	err = destFile.Sync()
	check(err)
}

func init() {
	rootCmd.AddCommand(setupCmd)

	setupCmd.PersistentFlags().StringVarP(&cliDBSetupUser, "user", "", "", "Custom database user")
	setupCmd.PersistentFlags().StringVarP(&cliDBSetupPassword, "password", "", "", "Custom database password")
}
