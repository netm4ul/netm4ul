package nmap

//package nmap

import (
	// "fmt"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/netm4ul/netm4ul/core/server/database"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/BurntSushi/toml"
	// gonmap "github.com/lair-framework/go-nmap"

	"github.com/netm4ul/netm4ul/modules"
)

//ConfigToml : configuration model (from the toml file)
type ConfigToml struct {
	FastScan   bool   `toml:"fast"`
	NoPing     bool   `toml:"noping"`
	Udp        bool   `toml:"udp"`
	PortRange  string `toml:"port_range"`
	Stealth    bool   `toml:"stealth"`
	Services   bool   `toml:"services"`
	OS         bool   `toml:"OS"`
	Verbose    bool   `toml:"verbose"`
	AllOptions bool   `toml:"all_options"`
	Ping       bool   `toml:"ping"`
}

// Nmap "class"
type Nmap struct {
	Config  ConfigToml
	Result  []byte
	Nmaprun *NmapRun
}

// NewTraceroute generate a new Nmap module (type modules.Module)
func NewNmap() modules.Module {

	gob.Register(NmapRun{})
	var t modules.Module
	t = &Nmap{}
	return t
}

// Name : name getter
func (N *Nmap) Name() string {
	return "Nmap"
}

// Author : Author getter
func (N *Nmap) Version() string {
	return "0.1"
}

// Version : Version getter
func (N *Nmap) Author() string {
	return "pruno"
}

// DependsOn : Generate the dependancies requirements
func (N *Nmap) DependsOn() []modules.Condition {
	var _ modules.Condition
	return []modules.Condition{}
}

// Run : Main function of the module
func (N *Nmap) Run(opt2 []string) (modules.Result, error) {
	N.ParseConfig()
	fmt.Println(&N.Config)
	var opt []string

	// Fast scan option : -F
	if N.Config.FastScan {
		opt = append(opt, "-F")
	}

	// Consider all hosts as up : -Pn
	if N.Config.NoPing {
		opt = append(opt, "-Pn")
	}

	// Ping scan option : -sP
	if N.Config.Ping {
		opt = append(opt, "-sP")
	}

	// UDP ports option : -sU
	if N.Config.Udp {
		opt = append(opt, "-sU")
	}

	// Port range option : -p- for all ports, -p x-y for specific range, nothing for default
	log.Println(N.Config.PortRange)
	if N.Config.PortRange != "Null" {
		opt = append(opt, "-p"+N.Config.PortRange)
	} else if N.Config.PortRange == "-" {
		opt = append(opt, "-p-")
	}

	// Stealth mode
	if N.Config.Stealth {
		opt = append(opt, "-sC")
	}

	// Service detection : -sV
	if N.Config.Services {
		opt = append(opt, "-sV")
	}

	// OS detection : -O
	if N.Config.OS {
		opt = append(opt, "-O")
	}

	// Verbose mode : -v
	if N.Config.Verbose {
		opt = append(opt, "-v")
	}

	// All options mode
	if N.Config.AllOptions {
		opt = append(opt, "-A")
	}

	// TODO : change it for per target option ?
	// filename := opt2[len(opt2)-1] + ".xml"
	filename := "127.0.0.1.xml"
	opt = append(opt, "-oX", filename)

	// TODO : change it for Run argument, will be passed as an option : ./netm4ul 127.0.0.1

	opt = append(opt, opt2...)

	fmt.Println(opt)
	cmd := exec.Command("/usr/bin/nmap", opt...)
	fmt.Println("My cmd:", cmd)
	execErr := cmd.Run()
	if execErr != nil {
		log.Fatal(execErr)
	}
	var err error
	N.Result, err = ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal("Error 2 : ", err)
	}
	N.Nmaprun, err = Parse(N.Result)
	return modules.Result{Data: N.Nmaprun, Timestamp: time.Now(), Module: N.Name()}, err
}

// HandleMQ : Recv data from the MQ
func (N *Nmap) HandleMQ() error {
	return nil
}

// SendMQ : Send data to the MQ
func (N *Nmap) SendMQ(data []byte) error {
	return nil
}

// ParseConfig : Load the config from the config folder
func (N *Nmap) ParseConfig() error {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}

	exPath := filepath.Dir(ex)
	configPath := filepath.Join(exPath, "config", "nmap.conf")
	_, err = toml.DecodeFile(configPath, &N.Config)

	if err != nil {
		log.Fatal("Error !", err)
		return err
	}

	return nil
}

// WriteDb : Save data
func (N *Nmap) WriteDb(result modules.Result, mgoSession *mgo.Session, projectName string) error {
	log.Println("Write to the database.")
	// var data NmapRun
	result.Data = result.Data.(NmapRun)
	// fmt.Printf("============================%+v", result.Data)
	//save raw data
	raw := bson.M{projectName + ".results." + result.Module: result}
	database.UpsertRawData(mgoSession, projectName, raw)

	//save data in projects

	return nil
}
