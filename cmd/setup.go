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
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/fatih/color"

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
	passwordLen        = 20
	defaultDBSetupUser = "postgres"
	defaultDBname      = "netm4ul"
	defaultDBIP        = "localhost"
	defaultDBPort      = uint16(5432)
	defaultDBType      = "postgresql"
	defaultAPIUser     = "user"
	defaultAPIIP       = "localhost"
	defaultAPIPort     = uint16(8080)
	defaultServerIP    = "localhost"
	defaultServerPort  = uint16(444)
	defaultServerUser  = "user"
	defaultTLS         = "y" // Yes/y No/n (case insensitive)
	defaultAlgorithm   = "random"
)

var (
	defaultDBSetupPassword, _ = GeneratePassword(passwordLen)
	defaultAPIPassword, _     = GeneratePassword(passwordLen)
	defaultServerPassword, _  = GeneratePassword(passwordLen)
	cliDBSetupUser            string
	cliDBSetupPassword        string
	cliDBSetupIP              string
	cliDBSetupPort            uint16
	cliDBSetupType            string
	cliServerUser             string
	cliServerPassword         string
	cliServerIP               string
	cliServerPort             uint16
	cliApiUser                string
	cliApiPassword            string
	cliApiIP                  string
	cliApiPort                uint16
	cliDisableTLS             bool
	cliAlgorithm              string
	skipDBSetup               bool
	skipServerSetup           bool
	skipApiSetup              bool
	skipAlgorithmSetup        bool
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

		if !skipAlgorithmSetup {
			err = setupAlgorithm()
			if err != nil {
				log.Fatalf("Could not setup the algorithm : %s", err.Error())
			}
		} else {
			fmt.Println("Skiping algorithm setup")
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
		"dbpassword":     {Message: "Database password (generated : (" + color.RedString("%s") + ") : ", DefaultValue: defaultDBSetupPassword},
		"dbip":           {Message: "Database IP (default : %s) : ", DefaultValue: defaultDBIP},
		"dbport":         {Message: "Database Port (default : %s) : ", DefaultValue: strconv.Itoa(int(defaultDBPort))},
		"dbtype":         {Message: "Database type [postgres, jsondb, mongodb] (default : %s): ", DefaultValue: defaultDBType},
		"apiuser":        {Message: "API username (default : %s) : ", DefaultValue: defaultAPIUser},
		"apipassword":    {Message: "API password (generated : (" + color.RedString("%s") + ") : ", DefaultValue: defaultAPIPassword},
		"apiport":        {Message: "API port (default : %s) : ", DefaultValue: strconv.Itoa(int(defaultAPIPort))},
		"serverip":       {Message: "Server IP (default : %s) : ", DefaultValue: defaultServerIP},
		"serverport":     {Message: "Server port (default : %s) : ", DefaultValue: strconv.Itoa(int(defaultServerPort))},
		"serveruser":     {Message: "Server username (default : %s) : ", DefaultValue: defaultServerUser},
		"serverpassword": {Message: "Server password (generated : (" + color.RedString("%s") + ") : ", DefaultValue: defaultServerPassword},
		"usetls":         {Message: "Use TLS (default : %s) [Y/n]: ", DefaultValue: defaultTLS},
		"algorithm":      {Message: "Load balancing algorithm (default : %s) : ", DefaultValue: defaultAlgorithm},
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
	tlsString := prompt("usetls")
	tlsBool, err := yesNo(tlsString)
	for err != nil {
		tlsString := prompt("usetls")
		tlsBool, err = yesNo(tlsString)
	}
	CLISession.Config.TLSParams.UseTLS = tlsBool
	return nil
}

func setupAlgorithm() error {
	var err error
	usedAlgo := prompt("algorithm")
	CLISession.Config.Algorithm.Name = usedAlgo

	return err
}
func yesNo(response string) (bool, error) {
	response = strings.ToLower(strings.TrimSpace(response))

	if response == "y" || response == "yes" {
		return true, nil
	} else if response == "n" || response == "no" {
		return false, nil
	}
	return false, errors.New("Invalid input")
}

// Save the config on disk
func saveConfigFile() error {

	//Create new config file
	file, err := os.OpenFile(
		CLISession.ConfigPath,
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
	cfgpath := CLISession.ConfigPath + ".example"
	data, err := ioutil.ReadFile(cfgpath)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(CLISession.ConfigPath, data, 0666)
	if err != nil {
		return err
	}

	return nil
}

//GeneratePassword returns a new password of len n (1st arg)
func GeneratePassword(n int) (string, error) {
	const codeAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789&#{}[]()-|_^@=+%?./!;,"
	pass := ""
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	for _, char := range b {
		pos := uint8(char)
		pass += string(codeAlphabet[int(pos)%len(codeAlphabet)])
	}

	return pass, nil
}

func init() {
	rootCmd.AddCommand(setupCmd)

	// database
	setupCmd.PersistentFlags().StringVar(&cliDBSetupUser, "database-user", defaultDBSetupUser, "Custom database user")
	setupCmd.PersistentFlags().StringVar(&cliDBSetupPassword, "database-password", defaultDBSetupPassword, "Custom database password")
	setupCmd.PersistentFlags().StringVar(&cliDBSetupIP, "database-ip", defaultDBIP, "Custom database ip address")
	setupCmd.PersistentFlags().Uint16Var(&cliDBSetupPort, "database-port", defaultDBPort, "Custom database port number")
	setupCmd.PersistentFlags().StringVar(&cliDBSetupType, "database-type", defaultDBType, "Custom database type number")
	//server
	setupCmd.PersistentFlags().StringVar(&cliServerUser, "server-user", defaultServerUser, "Custom server user")
	setupCmd.PersistentFlags().StringVar(&cliServerPassword, "server-password", defaultServerPassword, "Custom server password")
	setupCmd.PersistentFlags().StringVar(&cliServerIP, "server-ip", defaultServerIP, "Custom server ip address")
	setupCmd.PersistentFlags().Uint16Var(&cliServerPort, "server-port", defaultServerPort, "Custom server port number")
	//api
	setupCmd.PersistentFlags().StringVar(&cliApiUser, "api-user", defaultAPIUser, "Custom API user")
	setupCmd.PersistentFlags().StringVar(&cliApiPassword, "api-password", defaultAPIPassword, "Custom API password")
	setupCmd.PersistentFlags().StringVar(&cliApiIP, "api-ip", defaultAPIIP, "Custom API ip address")
	setupCmd.PersistentFlags().Uint16Var(&cliApiPort, "api-port", defaultAPIPort, "Custom API port number")
	//TLS
	setupCmd.PersistentFlags().BoolVar(&cliDisableTLS, "disable-tls", false, "Disable TLS")
	//Algorithm
	setupCmd.PersistentFlags().StringVar(&cliAlgorithm, "algorithm ", defaultAlgorithm, "Load balancing algorithm")
	//Skips
	setupCmd.PersistentFlags().BoolVar(&skipDBSetup, "skip-database", false, "Skip configuration of the database")
	setupCmd.PersistentFlags().BoolVar(&skipServerSetup, "skip-server", false, "Skip configuration of the server")
	setupCmd.PersistentFlags().BoolVar(&skipApiSetup, "skip-api", false, "Skip configuration of the Api")
	setupCmd.PersistentFlags().BoolVar(&skipAlgorithmSetup, "skip-algorithm", false, "Skip configuration of the algorithm")
}
