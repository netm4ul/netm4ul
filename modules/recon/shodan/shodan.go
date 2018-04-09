package shodan

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/netm4ul/netm4ul/modules"
	"os"
	"path/filepath"
	"log"
	"gopkg.in/ns3777k/go-shodan.v2/shodan"
)

// ConfigToml : configuration model (from the toml file)
type ConfigToml struct {
	// API_KEY int `toml:"api_key"`
}

// Shodan "class"
type ShodanResult struct {
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

func (S Shodan) NewShodan() modules.Module {
	gob.Register(ShodanResult{})
	var s modules.Module
	s = Shodan{}
	return s
}

// Run : Main function of the module
func (S Shodan) Run(data interface{}) (interface{}, error) {
	/*
		TODO: Not implemented yet
	*/
	fmt.Println("Shodan World!")

	// Get SHODAN_API_KEY from Environment Variables
	API_KEY = os.Getenv("SHODAN_API_KEY")
	fmt.Println(API_KEY)
	return nil, nil
}

// Parse : Parse the result of the execution
func (S Shodan) Parse() (interface{}, error) {
	return nil, nil
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
