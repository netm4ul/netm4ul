package shodan

import (
	"os"
	"path/filepath"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/netm4ul/netm4ul/modules"
)

// ConfigToml : configuration model (from the toml file)
type ConfigToml struct {
	API_KEY int `toml:"api_key"`
}

// Traceroute "class"
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
	return "Razbaa"
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

// Run : Main function of the module
func (S Shodan) Run(data interface{}) (interface{}, error) {
	/*
		TODO: Not implemented yet
	*/
	fmt.Println("NOT IMPLEMENTED YET")
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
