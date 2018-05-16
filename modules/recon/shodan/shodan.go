package shodan

import (
	"context"
	"encoding/gob"
	"os"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/BurntSushi/toml"
	"github.com/netm4ul/netm4ul/core/config"
<<<<<<< HEAD
	"github.com/netm4ul/netm4ul/core/database"
	"github.com/netm4ul/netm4ul/modules"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
=======
	"github.com/netm4ul/netm4ul/core/database/models"
	"github.com/netm4ul/netm4ul/modules"
>>>>>>> develop
	"gopkg.in/ns3777k/go-shodan.v3/shodan"
)

// ConfigToml : configuration model (from the toml file)
type ConfigToml struct {
	// API_KEY int `toml:"api_key"`
}

type ShodanResult struct {
	IP   string
	Host *shodan.Host
	// Services *Services
}

// Shodan "class"
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
	return "Rzbaa"
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

// NewShodan : Generate shodan object
func NewShodan() modules.Module {
	gob.Register(map[string]interface{}{})
	gob.Register([]interface{}{})
	gob.Register(ShodanResult{})
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

/*
	Usefull command
	curl -XPOST http://localhost:8080/api/v1/projects/FirstProject/run/shodan
	check db: Db.projects.find()
	remove all data: db.projects.remove({})
*/

// Run : Main function of the module
<<<<<<< HEAD
func (S Shodan) Run(data []string) (modules.Result, error) {
=======
func (S Shodan) Run(inputs []modules.Input) (modules.Result, error) {
>>>>>>> develop

	log.Debug("Shodan World!")

	// Instanciate shodanResult
	shodanResult := ShodanResult{}

	// Create client
	shodanClient := shodan.NewClient(nil, config.Config.Keys.Shodan)

	// Create shodan context
	shodanContext := context.Background()
	// Get IP adress
<<<<<<< HEAD
	dns, err := shodanClient.GetDNSResolve(shodanContext, []string{"google.com", "edznux.fr"})
	if err != nil {
		log.Panic(err)
	} else {
		// shodanResult.IP = *dns["edznux.fr"]
		myIP := *dns["edznux.fr"]
		shodanResult.IP = myIP.String()
	}
=======
	var domains []string
	for _, input := range inputs {
		if input.Domain != "" {
			domains = append(domains, input.Domain)
		}
	}
	dns, err := shodanClient.GetDNSResolve(shodanContext, domains)
	if err != nil {
		log.Panic(err)
	}
	// TODO : change shodanResult slices / array ?
	// Not sure about just one domain output...
	// shodanResult.IP = *dns["edznux.fr"]
	myIP := *dns[domains[0]]
	shodanResult.IP = myIP.String()
>>>>>>> develop

	hostServiceOption := shodan.HostServicesOptions{}

	// Get services of shodanResult.IP
	// log.Println(shodanResult.IP)
	host, err := shodanClient.GetServicesForHost(shodanContext, shodanResult.IP, &hostServiceOption)
	if err != nil {
		log.Panicln(err)
	}

	shodanResult.Host = host

	// for debug
	if config.Config.Verbose {
		printHost(*host)
		for _, servicesData := range host.Data {
			log.Debug(servicesData)
		}
	}

	// Exit message
	log.Debug("Shodan module executed. See u, in hell!!")

	return modules.Result{Data: shodanResult, Timestamp: time.Now(), Module: S.Name()}, err
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
func (S Shodan) ParseConfig() error {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	configPath := filepath.Join(exPath, "config", "shodan.conf")

	if _, err := toml.DecodeFile(configPath, &S.Config); err != nil {
		log.Error(err)
		return err
	}
	return nil
}

<<<<<<< HEAD
func (S Shodan) WriteDb(result modules.Result, mgoSession *mgo.Session, projectName string) error {
	log.Debug("Write to the database.")
	var data ShodanResult
	data = result.Data.(ShodanResult)

	raw := bson.M{projectName + ".results." + result.Module: data}
	database.UpsertRawData(mgoSession, projectName, raw)
=======
func (S Shodan) WriteDb(result modules.Result, db models.Database, projectName string) error {
	log.Debug("Write to the database.")
	// var data ShodanResult
	// data = result.Data.(ShodanResult)

	// raw := bson.M{projectName + ".results." + result.Module: data}
	// database.UpsertRawData(mgoSession, projectName, raw)
>>>>>>> develop
	return nil
}

//command: curl -XPOST http://localhost:8080/api/v1/projects/FirstProject/run/shodan
