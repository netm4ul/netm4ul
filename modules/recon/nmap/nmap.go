package nmap

//package nmap

import (
	"strings"
	// "fmt"
	"encoding/gob"
	"errors"
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
	FastScan       bool   `toml:"fast"`
	NoPing         bool   `toml:"no_ping"`
	UDP            bool   `toml:"udp"`
	PortRange      string `toml:"port_range"`
	Stealth        bool   `toml:"stealth"`
	Services       bool   `toml:"services"`
	OS             bool   `toml:"OS"`
	Verbose        bool   `toml:"verbose"`
	AllOptions     bool   `toml:"all_options"`
	Ping           bool   `toml:"ping"`
	TimingTemplate string `toml:"timing_template"`
	MinHostgroup   string `toml:"min_hostgroup"`
	MaxHostgroup   string `toml:"max_hostgroup"`
}

// Nmap "class"
type Nmap struct {
	Config  ConfigToml
	Result  []byte
	Nmaprun *gonmap.NmapRun
}

// NewNmap generate a new Nmap module (type modules.Module)
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

// Version : Version getter
func (N *Nmap) Version() string {
	return "0.1"
}

// Author : Author getter
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

	/*
	 We save the result in a temporary file (filename)
	 It's easier to get it parsed, have a backup and save it as raw
	*/
	opt, filename, err := N.loadArgs(inputs)
	cmd := exec.Command("/usr/bin/nmap", opt...)
	log.Debugf("Executing : %s %s", "/usr/bin/nmap", strings.Join(opt, " "))

	execErr := cmd.Run()
	if execErr != nil {
		return modules.Result{}, errors.New("Could not execute : " + execErr.Error())
	}
	/*
		The N.getResults read the file and get the raw and parsed output or error.
	*/
	N.Result, N.Nmaprun, err = N.getResults(filename)
	if err != nil {
		return modules.Result{}, errors.New("Could not get results : " + err.Error())
	}
	log.Debugf("Result  : %+v", N.Result)
	log.Debugf("Result parsed : %+v", N.Nmaprun)
	return modules.Result{Data: N.Nmaprun, Timestamp: time.Now(), Module: N.Name()}, err
}

func (N *Nmap) getResults(filename string) (raw []byte, parsed *gonmap.NmapRun, err error) {
	log.Debugf("Reading file : %s", filename)
	raw, err = ioutil.ReadFile(filename)
	if err != nil {
		return nil, nil, errors.New("Could not read file of nmap result : " + err.Error())
	}

	parsed, err = gonmap.Parse(raw)
	if err != nil {
		return nil, nil, errors.New("Could not parse nmap result : " + err.Error())
	}
	return raw, parsed, err
}

func (N *Nmap) loadArgs(inputs []modules.Input) (opt []string, filename string, err error) {
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
	if N.Config.PortRange != "" {
		opt = append(opt, "-p"+N.Config.PortRange)
	} else if N.Config.PortRange == "-" {
		opt = append(opt, "-p-")
	}

	// Stealth mode
	if N.Config.Stealth {
		opt = append(opt, "-sC")
	}

	// Timing template (-T3 is the nmap default, -T4 is recommended)
	if N.Config.TimingTemplate != "" {
		opt = append(opt, "-T"+N.Config.TimingTemplate)
	}

	if N.Config.MinHostgroup != "" {
		opt = append(opt, "--min-hostgroup="+N.Config.MinHostgroup)
	}

	if N.Config.MaxHostgroup != "" {
		opt = append(opt, "--min-hostgroup="+N.Config.MaxHostgroup)
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

	file, err := ioutil.TempFile("", "netm4ul_nmap_")
	if err != nil {
		return nil, "", errors.New("Could not create temp file for nmap result : " + err.Error())
	}
	defer os.Remove(file.Name())

	filename = file.Name()
	log.Debugf("Writing to file '%s'", filename)
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
	return opt, filename, nil
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
		log.Errorf("Couldn't parse nmap's config file : " + err.Error())
		return err
	}

	return nil
}

// WriteDb : Save data
func (N *Nmap) WriteDb(result modules.Result, db models.Database, projectName string) error {
	log.Info("Write raw results to the database.")

	// result.Data = result.Data.(NmapRun)
	data := result.Data.(gonmap.NmapRun)

	for _, host := range data.Hosts {
		// Ports part
		ports := make([]models.Port, len(host.Ports))
		for j := range data.Hosts[0].Ports {
			p := models.Port{
				Number:   int16(data.Hosts[0].Ports[j].PortId),
				Protocol: data.Hosts[0].Ports[j].Protocol,
				Status:   data.Hosts[0].Ports[j].State.State,
				Banner:   data.Hosts[0].Ports[j].Service.Name,
			}
			ports[j] = p
		}

		//TOFIX
		// Network external should be dynamicly updated
		for _, ip := range host.Addresses {
			element := models.IP{Value: ip.Addr, CreatedAt: time.Now(), UpdatedAt: time.Now(), Network: "external"}
			log.Debugf("Saving IP address : %+v", element)
			err := db.CreateOrUpdateIP(projectName, element)
			if err != nil {
				return errors.New("Could not save the database : " + err.Error())
			}
		}

		log.Debugf("Saving Ports : %+v for address %+v", ports, host.Addresses[0].Addr)
		err := db.CreateOrUpdatePorts(projectName, host.Addresses[0].Addr, ports)
		if err != nil {
			return errors.New("Could not save the database : " + err.Error())
		}
	}

	//save raw data
	now := time.Now()
	raw := models.Raw{
		Content:    string(N.Result),
		CreatedAt:  now,
		UpdatedAt:  now,
		ModuleName: N.Name(),
	}
	db.AppendRawData(projectName, raw)
	return nil
}
