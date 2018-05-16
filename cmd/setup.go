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
	defaultDBType          = "postgresql"
	defaultAPIUser         = "user"
	defaultAPIPassword     = "password"
	defaultAPIPort         = uint16(8080)
	defaultServerIP        = "localhost"
	defaultServerPort      = uint16(444)
	defaultServerUser      = "user"
	defaultServerPassword  = "password"
)

var (
	cliDBSetupUser     string
	cliDBSetupPassword string
	cliDBSetupIP       string
	cliDBSetupPort     uint16
	cliDBSetupType     string
	skipDBSetup        bool
	skipServerSetup    bool
	skipApiSetup       bool
)

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "NetM4ul setup",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		createSessionBase()
	},

	Run: func(cmd *cobra.Command, args []string) {
		err := copyExampleConf()
		if err != nil {
			log.Fatalf("Could not copy example file to standard config : %s", err.Error())
		}
		if !skipDBSetup {
			err = setupDB()
			if err != nil {
				log.Fatalf("Could not setup the database : %s", err.Error())
			}
		} else {
			fmt.Println("Skiping Database setup")
		}

		if !skipApiSetup {
			err = setupAPI()
			if err != nil {
				log.Fatalf("Could not setup the API : %s", err.Error())
			}
		} else {
			fmt.Println("Skiping API setup")
		}

		if !skipServerSetup {
			err = setupServer()
			if err != nil {
				log.Fatalf("Could not setup the Server : %s", err.Error())
			}
		} else {
			fmt.Println("Skiping Server setup")
		}

		err = saveConfigFile()
		if err != nil {
			log.Fatal("Could not save the file : " + err.Error())
		}
	},
}

// prompt user for configuration parameters
func prompt(param string) (answer string) {
	// var text string
	var input string

	// Database parameters
	promptString := map[string]PromptRes{
		"dbuser":         {Message: "Database username (default : %s) : ", DefaultValue: defaultDBSetupUser},
		"dbpassword":     {Message: "Database password (default : %s) : ", DefaultValue: defaultDBSetupPassword},
		"dbip":           {Message: "Database IP (default : %s) : ", DefaultValue: defaultDBIP},
		"dbport":         {Message: "Database Port (default : %s) : ", DefaultValue: strconv.Itoa(int(defaultDBPort))},
		"dbtype":         {Message: "Database type [postgres, jsondb, mongodb] (default : %s): ", DefaultValue: defaultDBType},
		"apiuser":        {Message: "API username (default : %s) : ", DefaultValue: defaultAPIUser},
		"apipassword":    {Message: "API password (default : %s) : ", DefaultValue: defaultAPIPassword},
		"apiport":        {Message: "API port (default : %s) : ", DefaultValue: strconv.Itoa(int(defaultAPIPort))},
		"serverip":       {Message: "Server IP (default : %s) : ", DefaultValue: defaultServerIP},
		"serverport":     {Message: "Server port (default : %s) : ", DefaultValue: strconv.Itoa(int(defaultServerPort))},
		"serveruser":     {Message: "Server username (default : %s) : ", DefaultValue: defaultServerUser},
		"serverpassword": {Message: "Server password (default : %s) : ", DefaultValue: defaultServerPassword},
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

	CLISession.Config.Database.IP = prompt("dbip")

	p, err := strconv.Atoi(prompt("dbport"))
	if err != nil {
		return err
	}
	CLISession.Config.Database.Port = uint16(p)
	CLISession.Config.Database.DatabaseType = prompt("dbtype")
	CLISession.Config.Database.User = prompt("dbuser")
	CLISession.Config.Database.Password = prompt("dbpassword")

	db := database.NewDatabase(&CLISession.Config)
	if db == nil {
		return errors.New("Could not create the database session")
	}

	err = db.SetupAuth(
		CLISession.Config.Database.User,
		CLISession.Config.Database.Password,
		CLISession.Config.Database.Database,
	)

	if err != nil {
		return errors.New("Could not setup the database : " + err.Error())
	}

	return nil
}

func setupAPI() error {
	CLISession.Config.API.User = prompt("apiuser")
	CLISession.Config.API.Password = prompt("apipassword")
	p, err := strconv.Atoi(prompt("apiport"))
	if err != nil {
		return err
	}
	CLISession.Config.API.Port = uint16(p)

	return nil
}

func setupServer() error {
	CLISession.Config.Server.IP = prompt("serverip")
	p, err := strconv.Atoi(prompt("serverport"))
	if err != nil {
		return err
	}
	CLISession.Config.Server.Port = uint16(p)
	CLISession.Config.Server.User = prompt("serveruser")
	CLISession.Config.Server.Password = prompt("serverpassword")

	return nil
}

// Save the config on disk
func saveConfigFile() error {

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
func copyExampleConf() error {
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
	setupCmd.PersistentFlags().BoolVar(&skipDBSetup, "skip-database", false, "Skip configuration of the database")
	setupCmd.PersistentFlags().BoolVar(&skipServerSetup, "skip-server", false, "Skip configuration of the server")
	setupCmd.PersistentFlags().BoolVar(&skipApiSetup, "skip-api", false, "Skip configuration of the Api")
}
