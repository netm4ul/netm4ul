package masscan

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
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
	Raw      string
	Resultat Scan
}

// Scan represents the ip and ports output
type Scan struct {
	IP    string
	Ports Port
}

// Port represents the port, proto, service, ttl, reason and status output
type Port struct {
	Port    uint16
	Proto   string
	Service Service
	TTL     int
	Reason  string
	Status  string
}

// Service represents the name and the banner output
type Service struct {
	Name   string
	Banner string
}

// ConfigToml : configuration model (from the toml file)
type ConfigToml struct {
	Ports   string `toml:"ports"`
	Banner  bool   `toml:"banner"`
	Source  string `toml:"source"`
	Rate    string `toml:"rate"`
	Verbose bool   `toml:"verbose"`
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
	t = &Masscan{}
	return t
}

// Name : name getter
func (M *Masscan) Name() string {
	return "Masscan"
}

// Author : Author getter
func (M *Masscan) Author() string {
	return "soldat-ryan"
}

// Version : Version  getter
func (M *Masscan) Version() string {
	return "0.1"
}

// DependsOn : Generate the dependencies requirement
func (M *Masscan) DependsOn() []modules.Condition {
	var _ modules.Condition
	return []modules.Condition{}
}

// Checks error
func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

// Run : Main function of the module
func (M *Masscan) Run(data []string) (modules.Result, error) {
	var opt []string
	outputfile := "output.json"

	fmt.Println("hello world masscan") //Affiche hello world pour le fun
	M.ParseConfig()

	log.Printf("Verbose mode: %+v", M.Config.Verbose)
	if M.Config.Verbose {
		opt = append(opt, "-v")
	}

	// IP forced : 212.47.247.190 = edznux.fr
	//opt = append(opt, data...)
	opt = append(opt, "212.47.247.190")

	// Ports option
	log.Println(M.Config.Ports)
	if M.Config.Ports != "" {
		opt = append(opt, "-p"+M.Config.Ports)
	} else {
		opt = append(opt, "-p1-65535")
	}

	// Banner option
	log.Println(M.Config.Banner)
	if M.Config.Banner {
		opt = append(opt, "--banners")
	}

	// Rate option
	log.Println(M.Config.Rate)
	if M.Config.Rate != "" {
		opt = append(opt, "--rate="+M.Config.Rate)
	}

	// Output option
	opt = append(opt, "-oJ", outputfile)

	cmd := exec.Command("masscan", opt...)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	err := cmd.Run()
	check(err)

	content, err := ioutil.ReadFile(outputfile)
	check(err)
	fmt.Printf("data: %s", string(content))

	return modules.Result{Data: MasscanResult{Raw: string(content)}, Timestamp: time.Now(), Module: M.Name()}, nil
}

// Parse : Parse the result of the execution
func (M *Masscan) Parse() (MasscanResult, error) {
	return MasscanResult{}, nil
}

// ParseConfig : Load the config from the config folder
func (M *Masscan) ParseConfig() error {
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
func (M *Masscan) WriteDb(result modules.Result, mgoSession *mgo.Session, projectName string) error {
	log.Println("Write to the database.")
	var data MasscanResult
	data = result.Data.(MasscanResult)

	raw := bson.M{projectName + ".results." + result.Module: data}
	database.UpsertRawData(mgoSession, projectName, raw)
	return nil
}
