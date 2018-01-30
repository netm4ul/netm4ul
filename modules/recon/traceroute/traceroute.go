package traceroute

import (
	"fmt"
	//"github.com/BurntSushi/toml"
	//"github.com/netm4ul/netm4ul/modules"
	"bytes"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/netm4ul/netm4ul/modules"
)

// ConfigToml : configuration model (from the toml file)
type ConfigToml struct {
	MaxHops int `toml:"max_hops"`
}

// Traceroute "class"
type Traceroute struct {
	// Config : exported config
	Config ConfigToml
}

// Name : name getter
func (T *Traceroute) Name() string {
	return "Traceroute"
}

// Author : Author getter
func (T *Traceroute) Author() string {
	return "tomalavie"
}

// Version : Version  getter
func (T *Traceroute) Version() string {
	return "0.1"
}

// DependsOn : Generate the dependencies requirement
func (T *Traceroute) DependsOn() []modules.Condition {
	var _ modules.Condition
	return []modules.Condition{}
}

// Run : Main function of the module
func (T *Traceroute) Run(data interface{}) (interface{}, error) {
	fmt.Println("hello world")                   //Affiche hello world pour le fun
	cmd := exec.Command("traceroute", "8.8.8.8") //
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf(out.String())
	return nil, nil
}

// Parse : Parse the result of the execution
func (T *Traceroute) Parse() (interface{}, error) {
	return nil, nil
}

// HandleMQ : Recv data from the MQ
func (T *Traceroute) HandleMQ() error {
	return nil
}

// SendMQ : Send data to the MQ
func (T *Traceroute) SendMQ(data []byte) error {
	return nil
}

// ParseConfig : Load the config from the config folder
func (T *Traceroute) ParseConfig() error {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	configPath := filepath.Join(exPath, "config", "traceroute.conf")

	if _, err := toml.DecodeFile(configPath, &T.Config); err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(T.Config.MaxHops)
	return nil
}
