package nmap

//package nmap

import (
	// "fmt"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/BurntSushi/toml"
	gonmap "github.com/edznux/go-nmap"

	"github.com/netm4ul/netm4ul/core/database/models"
	"github.com/netm4ul/netm4ul/modules"
)

//ConfigToml : configuration model (from the toml file)
type ConfigToml struct {
	FastScan   bool   `toml:"fast"`
	NoPing     bool   `toml:"no_ping"`
	UDP        bool   `toml:"udp"`
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
	Nmaprun *gonmap.NmapRun
}

// NewTraceroute generate a new Nmap module (type modules.Module)
func NewNmap() modules.Module {

	gob.Register(gonmap.NmapRun{})
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
func (N *Nmap) Run(inputs []modules.Input) (modules.Result, error) {
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
	if N.Config.UDP {
		opt = append(opt, "-sU")
	}

	// Port range option : -p- for all ports, -p x-y for specific range, nothing for default
	// TODO : load ports from []inputs.Ports ?
	log.Infof("PortRange : %+v", N.Config.PortRange)
	if N.Config.PortRange != "NULL" {
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
	for _, input := range inputs {
		if input.Domain != "" {
			opt = append(opt, input.Domain)
		}
		if input.IP != nil {
			opt = append(opt, input.IP.String())
		}
	}

	fmt.Println(opt)
	cmd := exec.Command("/usr/bin/nmap", opt...)
	execErr := cmd.Run()
	if execErr != nil {
		log.Fatalf("Could not execute : %+v ", execErr)
	}
	var err error
	N.Result, err = ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal("Error 2 : ", err)
	}
	N.Nmaprun, err = gonmap.Parse(N.Result)
	return modules.Result{Data: N.Nmaprun, Timestamp: time.Now(), Module: N.Name()}, err
}

// ParseConfig : Load the config from the config folder
func (N *Nmap) ParseConfig() error {
	ex, err := os.Executable()
	if err != nil {
		log.Fatal(err)
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
func (N *Nmap) WriteDb(result modules.Result, db models.Database, projectName string) error {
	log.Println("Write raw to the database.")

	// result.Data = result.Data.(NmapRun)
	data := result.Data.(gonmap.NmapRun)

	//save data in projects

	// define infos to send in db
	// Ports part
	ports := make([]models.Port, len(data.Hosts[0].Ports))
	for j := range data.Hosts[0].Ports {
		p := models.Port{
			Number:   int16(data.Hosts[0].Ports[j].PortId),
			Protocol: data.Hosts[0].Ports[j].Protocol,
			Status:   data.Hosts[0].Ports[j].State.State,
			Banner:   data.Hosts[0].Ports[j].Service.Name,
		}
		ports[j] = p
	}

	// IP parts, multi ips ?
	// var targets []database.IP
	// for i := range data.Hosts {
	// 	targets[i].Value = net.ParseIP(data.Hosts[0].Addresses[i].Addr)
	// 	targets[i].Ports = ports
	// }

	// For now, only 1 IP
	var target models.IP
	target.Value = data.Hosts[0].Addresses[0].Addr
	target.Ports = ports

	// put everything in db
	//IP
	//change to CreateOrUpdateIPs
	db.CreateOrUpdateIP(projectName, target)
	// Ports
	db.CreateOrUpdatePorts(projectName, target.Value, ports)
	//save raw data
	db.AppendRawData(projectName, result.Module, data)
	return nil
}
