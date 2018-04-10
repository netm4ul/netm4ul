package masscan

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"os/exec"
	"time"

	"os"
	"path/filepath"

	"github.com/netm4ul/netm4ul/core/server/database"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/BurntSushi/toml"
	"github.com/netm4ul/netm4ul/modules"
)

// MasscanResult represent the parsed ouput
type MasscanResult struct {
	IP    string
	Ports int
}

// ConfigToml : configuration model (from the toml file)
type ConfigToml struct {
	//MaxHops int `toml:"max_hops"`
	Ports  int    `toml:"ports"`
	Output string `toml:"output"`
	Rate   int    `toml:"rate"`
	Banner string `toml:"banner"`
}

// Masscan "class"
type Masscan struct {
	// Config : exported config
	Config ConfigToml
}

//NewMasscan generate a new Masscan module (type modules.Module)
func NewMasscan() modules.Module {
	gob.Register(MasscanResult{})
	var t modules.Module
	t = Masscan{}
	return t
}

// Name : name getter
func (M Masscan) Name() string {
	return "Masscan"
}

// Author : Author getter
func (M Masscan) Author() string {
	return "soldat-ryan"
}

// Version : Version  getter
func (M Masscan) Version() string {
	return "0.1"
}

// DependsOn : Generate the dependencies requirement
func (M Masscan) DependsOn() []modules.Condition {
	var _ modules.Condition
	return []modules.Condition{}
}

// Run : Main function of the module
func (M Masscan) Run(data []string) (modules.Result, error) {
	fmt.Println("hello world") //Affiche hello world pour le fun
	// "212.47.247.190" = edznux.fr
	cmd := exec.Command("sudo", "masscan", "212.47.247.190", "-p80")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf(out.String())
	return modules.Result{Data: MasscanResult{IP: "212.47.247.190", Ports: 80}, Timestamp: time.Now(), Module: M.Name()}, nil
}

// Parse : Parse the result of the execution
func (M Masscan) Parse() (MasscanResult, error) {
	return MasscanResult{}, nil
}

// ParseConfig : Load the config from the config folder
func (M Masscan) ParseConfig() error {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	configPath := filepath.Join(exPath, "config", "masscan.conf")

	if _, err := toml.DecodeFile(configPath, &M.Config); err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

// WriteDb : Save data
func (M Masscan) WriteDb(result modules.Result, mgoSession *mgo.Session, projectName string) error {
	log.Println("Write to the database.")
	var data MasscanResult
	data = result.Data.(MasscanResult)

	raw := bson.M{projectName + ".results." + result.Module: data}
	database.UpsertRawData(mgoSession, projectName, raw)
	return nil
}
