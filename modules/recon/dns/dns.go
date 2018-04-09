package dns

import (
	"encoding/gob"
	"fmt"
	"log"
	"time"

	"os"
	"path/filepath"

	"github.com/netm4ul/netm4ul/cmd/server/database"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/BurntSushi/toml"
	"github.com/netm4ul/netm4ul/modules"
)

// DnsResult represent the parsed ouput
type DnsResult struct {
	Test string
}

// ConfigToml : configuration model (from the toml file)
type ConfigToml struct {
	MaxHops int `toml:"max_hops"`
}

// Dns "class"
type Dns struct {
	// Config : exported config
	Config ConfigToml
}

//NewDns generate a new Dns module (type modules.Module)
func NewDns() modules.Module {
	gob.Register(DnsResult{})
	var t modules.Module
	t = Dns{}
	return t
}

// Name : name getter
func (T Dns) Name() string {
	return "Dns"
}

// Author : Author getter
func (T Dns) Author() string {
	return "tomalavie"
}

// Version : Version  getter
func (T Dns) Version() string {
	return "0.1"
}

// DependsOn : Generate the dependencies requirement
func (T Dns) DependsOn() []modules.Condition {
	var _ modules.Condition
	return []modules.Condition{}
}

// Run : Main function of the module
func (T Dns) Run(data []string) (modules.Result, error) {
	fmt.Println("DNS world!")
	return modules.Result{Data: DnsResult{Test: "Zgeg"}, Timestamp: time.Now(), Module: T.Name()}, nil
}

// ParseConfig : Load the config from the config folder
func (T Dns) ParseConfig() error {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	configPath := filepath.Join(exPath, "config", "dns.conf")

	if _, err := toml.DecodeFile(configPath, &T.Config); err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(T.Config.MaxHops)
	return nil
}

// WriteDb : Save data
func (T Dns) WriteDb(result modules.Result, mgoSession *mgo.Session, projectName string) error {
	log.Println("Write to the database.")
	var data DnsResult
	data = result.Data.(DnsResult)

	raw := bson.M{projectName + ".results." + result.Module: data}
	database.UpsertRawData(mgoSession, projectName, raw)
	return nil
}
