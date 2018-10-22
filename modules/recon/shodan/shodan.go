package shodan

import (
	"context"
	"encoding/gob"
	"errors"
	"os"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/BurntSushi/toml"
	"github.com/netm4ul/netm4ul/core/communication"
	"github.com/netm4ul/netm4ul/core/config"
	"github.com/netm4ul/netm4ul/core/database/models"
	"github.com/netm4ul/netm4ul/modules"
	"gopkg.in/ns3777k/go-shodan.v3/shodan"
)

// ConfigShodan : configuration model (from the toml file)
type ConfigShodan struct {
	// API_KEY int `toml:"api_key"`
}

//Result represent the data structure for the shodan module
type Result struct {
	IP   string
	Host *shodan.Host
	// Services *Services
}

// Shodan "class"
type Shodan struct {
	// Config : exported config
	ConfigShodan ConfigShodan
	Config       config.ConfigToml
}

// Name : name getter
func (S *Shodan) Name() string {
	return "Shodan"
}

// Author : Author getter
func (S *Shodan) Author() string {
	return "Rzbaa"
}

// Version : Version  getter
func (S *Shodan) Version() string {
	return "0.1"
}

// DependsOn : Generate the dependencies requirement
func (S *Shodan) DependsOn() []modules.Condition {
	var _ modules.Condition
	return []modules.Condition{}
}

// NewShodan : Generate shodan object
func NewShodan() modules.Module {
	gob.Register(map[string]interface{}{})
	gob.Register([]interface{}{})
	gob.Register(Result{})
	var s modules.Module
	s = &Shodan{}
	return s
}

// Parsing
/*
type Services struct {
	Product      string
	Organization string
	Data         string
	ASN          string
	Port         int
	Location     *shodan.HostLocation
}
*/

// Run : Main function of the module
func (S *Shodan) Run(input communication.Input, resultChan chan communication.Result) (communication.Done, error) {

	// Instanciate Result
	Result := Result{}
	// Create client
	shodanClient := shodan.NewClient(nil, S.Config.Keys.Shodan)
	// Create shodan context
	shodanContext := context.Background()
	// Get IP address
	var domain string

	if input.Domain == "" {
		err := errors.New("Empty domain provided, can't run shodan")
		return communication.Done{Error: err}, err
	}

	dns, err := shodanClient.GetDNSResolve(shodanContext, []string{input.Domain})
	if err != nil {
		return communication.Done{}, err
	}
	myIP := *dns[domain]
	Result.IP = myIP.String()

	hostServiceOption := shodan.HostServicesOptions{}

	// Get services of Result.IP
	// log.Println(Result.IP)
	host, err := shodanClient.GetServicesForHost(shodanContext, Result.IP, &hostServiceOption)
	if err != nil {
		return communication.Done{Error: err}, err
	}

	Result.Host = host

	printHost(*host)
	for _, servicesData := range host.Data {
		log.Debug(servicesData)
	}

	resultChan <- communication.Result{Data: Result, Timestamp: time.Now(), ModuleName: S.Name()}

	return communication.Done{Timestamp: time.Now(), ModuleName: S.Name()}, nil
}

func printHost(host shodan.Host) {
	log.Debug(host.OS)
	log.Debug(host.Ports)
	log.Debug(host.IP)
	log.Debug(host.ISP)
	log.Debug(host.Hostnames)
	log.Debug(host.Organization)
	log.Debug(host.Vulnerabilities)
	log.Debug(host.ASN)
	log.Debug(host.LastUpdate)
	log.Debug(host.Data)
	log.Debug(host.HostLocation)
}

// ParseConfig : Load the config from the config folder
func (S *Shodan) ParseConfig() error {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	configPath := filepath.Join(exPath, "config", "shodan.conf")

	if _, err := toml.DecodeFile(configPath, &S.ConfigShodan); err != nil {
		log.Error(err)
		return err
	}
	return nil
}
//WriteDb saves the information into the database
func (S *Shodan) WriteDb(result communication.Result, db models.Database, projectName string) error {
	log.Debug("Write to the database.")
	// var data Result
	// data = result.Data.(Result)

	// raw := bson.M{projectName + ".results." + result.Module: data}
	// database.UpsertRawData(mgoSession, projectName, raw)
	return nil
}

//command: curl -XPOST http://localhost:8080/api/v1/projects/FirstProject/run/shodan
