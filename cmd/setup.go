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
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/BurntSushi/toml"
	"github.com/netm4ul/netm4ul/core/database"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type PromptRes struct {
	Message      string
	DefaultValue string
}

const (
	defaultDBSetupUser     = "postgres"
	defaultDBSetupPassword = "password"
	defaultDBname          = "netm4ul"
	defaultDBIP            = "localhost"
	defaultDBPort          = uint16(5432)
	defaultDBType          = "postgres"
)

var (
	cliDBSetupUser     string
	cliDBSetupPassword string
	cliDBSetupIP       string
	cliDBSetupPort     uint16
	cliDBSetupType     string
)

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "NetM4ul setup",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
	},

	Run: func(cmd *cobra.Command, args []string) {
		err := example2Conf()
		if err != nil {
			log.Fatalf("Could not copy example file to standard config : %s", err.Error())
		}

		err = setupDB()
		if err != nil {
			log.Fatalf("Could not setup the database : %s", err.Error())
		}

	},
}

// prompt user for configuration parameters
func prompt(param string) (answer string) {
	// var text string
	var input string

	// Database parameters
	promptString := map[string]PromptRes{
		"dbuser":     {Message: "Database username (default : %s) : ", DefaultValue: defaultDBSetupUser},
		"dbpassword": {Message: "Database password (default : %s) : ", DefaultValue: defaultDBSetupPassword},
		"dbip":       {Message: "Database IP (default : %s) : ", DefaultValue: defaultDBIP},
		"dbport":     {Message: "Database Port (default : %s) : ", DefaultValue: strconv.Itoa(int(defaultDBPort))},
		"dbtype":     {Message: "Database type [postgres, jsondb, mongodb] (default : %s): ", DefaultValue: defaultDBType},
	}

	fmt.Printf(promptString[param].Message, promptString[param].DefaultValue)
	fmt.Scanln(&input)
	if input == "" {
		return promptString[param].DefaultValue
	}

	return input
}

// setupDB => create user in db for future requests
func setupDB() error {

	var err error
	db := database.NewDatabase(&CLISession.Config)

	if CLISession.Config.Database.IP == defaultDBIP {
		CLISession.Config.Database.IP = prompt("dbip")
	}

	if CLISession.Config.Database.Port == defaultDBPort {

		p, err := strconv.Atoi(prompt("dbport"))
		if err != nil {
			return err
		}

		CLISession.Config.Database.Port = uint16(p)
	}
	if CLISession.Config.Database.DatabaseType == defaultDBType {
		CLISession.Config.Database.DatabaseType = prompt("dbtype")
	}

	if CLISession.Config.Database.User == defaultDBSetupUser {
		CLISession.Config.Database.User = prompt("dbuser")
	}

	if CLISession.Config.Database.Password == defaultDBSetupPassword {
		CLISession.Config.Database.Password = prompt("dbpassword")
	}

	err = db.SetupAuth(
		CLISession.Config.Database.User,
		CLISession.Config.Database.Password,
		defaultDBname,
	)

	if err != nil {
		return errors.New("Could not setup the database : " + err.Error())
	}

	err = modifyDBConnect()
	if err != nil {
		return errors.New("Could not save the file : " + err.Error())
	}
	return nil
}

// modify conf file for db parameters
func modifyDBConnect() error {

	//Create new config file
	file, err := os.OpenFile(
		CLISession.Config.ConfigPath,
		os.O_WRONLY|os.O_TRUNC|os.O_CREATE,
		0666,
	)

	if err != nil {
		return err
	}

	defer file.Close()

	//Write new config to new config file
	err = toml.NewEncoder(file).Encode(CLISession.Config)
	if err != nil {
		return err
	}

	return nil
}

// modify conf file
func example2Conf() error {
	cfgpath := CLISession.Config.ConfigPath + ".example"
	data, err := ioutil.ReadFile(cfgpath)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(CLISession.Config.ConfigPath, data, 0666)
	if err != nil {
		return err
	}

	return nil
}

func init() {
	rootCmd.AddCommand(setupCmd)

	setupCmd.PersistentFlags().StringVar(&cliDBSetupUser, "user", defaultDBSetupUser, "Custom database user")
	setupCmd.PersistentFlags().StringVar(&cliDBSetupPassword, "password", defaultDBSetupPassword, "Custom database password")
	setupCmd.PersistentFlags().StringVar(&cliDBSetupIP, "database-ip", defaultDBIP, "Custom database ip address")
	setupCmd.PersistentFlags().Uint16Var(&cliDBSetupPort, "database-port", defaultDBPort, "Custom database port number")
	setupCmd.PersistentFlags().StringVar(&cliDBSetupType, "database-type", defaultDBType, "Custom database type")
}
