package shodan

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/netm4ul/netm4ul/cmd/config"
	"github.com/netm4ul/netm4ul/cmd/server/database"
	"github.com/netm4ul/netm4ul/modules"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/ns3777k/go-shodan.v2/shodan"
)

// ConfigToml : configuration model (from the toml file)
type ConfigToml struct {
	// API_KEY int `toml:"api_key"`
}

type ShodanResult struct {
	Test string
}

// Shodan "class"
type Shodan struct {
	// Config : exported config
	Config ConfigToml
}

// Name : name getter
func (S Shodan) Name() string {
	return "Shodan"
}

// Author : Author getter
func (S Shodan) Author() string {
	return "Rzbaa"
}

// Version : Version  getter
func (S Shodan) Version() string {
	return "0.1"
}

// DependsOn : Generate the dependencies requirement
func (S Shodan) DependsOn() []modules.Condition {
	var _ modules.Condition
	return []modules.Condition{}
}

// NewShodan : Generate shodan object
func NewShodan() modules.Module {
	gob.Register(ShodanResult{})
	var s modules.Module
	s = Shodan{}
	return s
}

// Run : Main function of the module
func (S Shodan) Run(data []string) (modules.Result, error) {

	/*
		TODO: Not implemented yet
	*/

	fmt.Println("Shodan World!")

	client := shodan.NewClient(nil, config.Config.Keys.Shodan)
	dns, err := client.GetDNSResolve([]string{"google.com", "edznux.fr"})

	if err != nil {
		log.Panic(err)
	} else {
		log.Println(*dns["edznux.fr"])
	}

	return modules.Result{Data: ShodanResult{Test: "Zgeg"}, Timestamp: time.Now(), Module: S.Name()}, nil
}

// HandleMQ : Recv data from the MQ
func (S Shodan) HandleMQ() error {
	return nil
}

// SendMQ : Send data to the MQ
func (S Shodan) SendMQ(data []byte) error {
	return nil
}

// ParseConfig : Load the config from the config folder
func (S Shodan) ParseConfig() error {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	configPath := filepath.Join(exPath, "config", "shodan.conf")

	if _, err := toml.DecodeFile(configPath, &S.Config); err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (S Shodan) WriteDb(result modules.Result, mgoSession *mgo.Session, projectName string) error {
	log.Println("Write to the database.")
	var data ShodanResult
	data = result.Data.(ShodanResult)

	raw := bson.M{projectName + ".results." + result.Module: data}
	database.UpsertRawData(mgoSession, projectName, raw)
	return nil
}

//command: curl -XPOST http://localhost:8080/api/v1/projects/FirstProject/run/shodan
