package masscan

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"regexp"
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
	Resultat []Scan
}

// Scan represents the ip and ports output
type Scan struct {
	IP    string `json:"ip"`
	Ports []Port `json:"ports"`
}

// Port represents the port, proto, service, ttl, reason and status output
type Port struct {
	Port    uint16  `json:"port"`
	Proto   string  `json:"proto"`
	Service Service `json:"service,omitempty"`
	TTL     int     `json:"ttl"`
	Reason  string  `json:"reason"`
	Status  string  `json:"status"`
}

// Service represents the name and the banner output
type Service struct {
	Name   string `json:"name"`
	Banner string `json:"banner"`
}

// ConfigToml : configuration model (from the toml file)
type ConfigToml struct {
	Ports             string `toml:"ports"`
	Banner            bool   `toml:"banner"`
	ConnectionTimeout int    `toml:"connection-timeout"`
	Rate              string `toml:"rate"`
	Verbose           bool   `toml:"verbose"`
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
	fmt.Println("H3ll-0 M4sscan")

	outputfile := "output.json"
	opt := M.ParseOptions()
	opt = append(opt, "-oJ", outputfile)

	// IP forced : 212.47.247.190 = edznux.fr
	//opt = append(opt, data...)
	opt = append([]string{"212.47.247.190"}, opt...)

	cmd := exec.Command("masscan", opt...)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	err := cmd.Run()
	check(err)

	res, err := M.Parse(outputfile)

	return modules.Result{Data: res, Timestamp: time.Now(), Module: M.Name()}, nil
}

// ParseOptions : Parse the args in according to masscan.conf
func (M *Masscan) ParseOptions() []string {
	var opt []string

	M.ParseConfig()

	// Verbose option
	if M.Config.Verbose {
		opt = append(opt, "-v")
	}
	// Ports option
	if M.Config.Ports != "" {
		opt = append(opt, "-p"+M.Config.Ports)
	} else {
		opt = append(opt, "-p1-65535")
	}
	// Banner option
	if M.Config.Banner {
		opt = append(opt, "--banners")
	}
	// Connection-time option
	if M.Config.ConnectionTimeout != 0 {
		opt = append(opt, "--connection-timeout", string(M.Config.ConnectionTimeout))
	}
	// Rate option
	if M.Config.Rate != "" {
		opt = append(opt, "--rate="+M.Config.Rate)
	}

	return opt
}

// Parse : Parse the result of the execution
func (M *Masscan) Parse(file string) (MasscanResult, error) {
	var scans []Scan

	data, err := ioutil.ReadFile(file)
	check(err)

	// JSON reformatted
	re := regexp.MustCompile(",\n{finished:.*}")
	fileReformatted := "[" + re.ReplaceAllString(string(data), "]")

	err = json.Unmarshal([]byte(fileReformatted), &scans)
	check(err)

	return MasscanResult{Resultat: scans}, nil
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
