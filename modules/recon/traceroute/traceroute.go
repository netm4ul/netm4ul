package traceroute

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

// TracerouteResult represent the parsed ouput
type TracerouteResult struct {
	Source      string
	Destination string
	Max         float32
	Min         float32
	Avg         float32
}

// ConfigToml : configuration model (from the toml file)
type ConfigToml struct {
	MaxHops int `toml:"max_hops"`
}

// Traceroute "class"
type Traceroute struct {
	// Config : exported config
	Config ConfigToml
}

//NewTraceroute generate a new Traceroute module (type modules.Module)
func NewTraceroute() modules.Module {
	gob.Register(TracerouteResult{})
	var t modules.Module
	t = &Traceroute{}
	return t
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
func (T *Traceroute) Run(data []string) (modules.Result, error) {
	fmt.Println("hello world") //Affiche hello world pour le fun
	// cmd := exec.Command("traceroute", "8.8.8.8") //
	// var out bytes.Buffer
	// cmd.Stdout = &out
	// err := cmd.Run()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Printf(out.String())
	return modules.Result{Data: TracerouteResult{Source: "SRC", Destination: "DST", Min: 12.3, Max: 123.4, Avg: 56.78}, Timestamp: time.Now(), Module: T.Name()}, nil
}

// Parse : Parse the result of the execution
func (T *Traceroute) Parse() (TracerouteResult, error) {
	return TracerouteResult{}, nil
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

// WriteDb : Save data
func (T *Traceroute) WriteDb(result modules.Result, mgoSession *mgo.Session, projectName string) error {
	log.Println("Write to the database.")
	var data TracerouteResult
	data = result.Data.(TracerouteResult)

	raw := bson.M{projectName + ".results." + result.Module: data}
	database.UpsertRawData(mgoSession, projectName, raw)
	return nil
}
