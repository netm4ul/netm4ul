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
	"fmt"
	"log"

	"github.com/BurntSushi/toml"
	"github.com/netm4ul/netm4ul/core/config"
	"github.com/netm4ul/netm4ul/core/server/database"
	"github.com/spf13/cobra"

	mgo "gopkg.in/mgo.v2"
)

const (
	defaultSetupUser     = "admin"
	defaultSetupPassword = "admin"
	dbname               = "NetM4ul"
)

var (
	setupUser     string
	setupPassword string
)

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "NetM4ul setup",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		setupDB()
	},
}

// setupDB => create user in db for future requests
func setupDB() {

	fmt.Println("Configuring the database with provided user")

	mgoSession := database.ConnectWithoutCreds()
	roles := []mgo.Role{mgo.RoleDBAdmin}

	u := mgo.User{Username: setupUser, Password: setupPassword, Roles: roles}

	c := mgoSession.DB(dbname)

	err := c.UpsertUser(&u)

	if err != nil {
		log.Fatal(err)
	}
	modifyDBConnect()
}

func modifyDBConnect() {
	config.Config.Database.User = setupUser
	config.Config.Database.Password = setupPassword

	fmt.Println(config.Config)

	buf := new(bytes.Buffer)
	if err := toml.NewEncoder(buf).Encode(config.Config); err != nil {
		log.Fatal(err)
	}
	fmt.Println(buf.String())

}

func init() {
	rootCmd.AddCommand(setupCmd)
	setupCmd.PersistentFlags().StringVarP(&setupUser, "user", "u", defaultSetupUser, "Custom database user")
	setupCmd.PersistentFlags().StringVarP(&setupPassword, "password", "p", defaultSetupPassword, "Custom database password")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// setupCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// setupCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
