package shodan

import (
	"context"
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/netm4ul/netm4ul/core/config"
	"github.com/netm4ul/netm4ul/core/server/database"
	"github.com/netm4ul/netm4ul/modules"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/ns3777k/go-shodan.v3/shodan"
)

// ConfigToml : configuration model (from the toml file)
type ConfigToml struct {
	// API_KEY int `toml:"api_key"`
}

type ShodanResult struct {
	IP    string
	Ports []int
	Host  *shodan.Host
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
	gob.Register(ShodanResult{})
	var s modules.Module
	s = Shodan{}
	return s
}

/*
	Usefull command
	curl -XPOST http://localhost:8080/api/v1/projects/FirstProject/run/shodan
	check db: Db.projects.find()
	remove all data: db.projects.remove({})
*/

// Run : Main function of the module
func (S Shodan) Run(data []string) (modules.Result, error) {

	/*
		TODO: Not implemented yet
	*/

	fmt.Println("Shodan World!")

	// Instanciate shodanResult
	shodanResult := ShodanResult{}

	// Create client
	shodanClient := shodan.NewClient(nil, config.Config.Keys.Shodan)

	// Get IP adress
	shodanContext := context.Background()
	dns, err := shodanClient.GetDNSResolve(shodanContext, []string{"google.com", "edznux.fr"})
	if err != nil {
		log.Panic(err)
	} else {
		// shodanResult.IP = *dns["edznux.fr"]
		myIP := *dns["edznux.fr"]
		shodanResult.IP = myIP.String()
	}

	hostServiceOption := shodan.HostServicesOptions{}

	// Get services of shodanResult.IP
	// log.Println(shodanResult.IP)
	host, err := shodanClient.GetServicesForHost(shodanContext, shodanResult.IP, &hostServiceOption)
	if err != nil {
		log.Panicln(err)
	}
	shodanResult.Host = host
	printHost(*host)
	return modules.Result{Data: shodanResult, Timestamp: time.Now(), Module: S.Name()}, err
}

func printHost(host shodan.Host) {
	log.Println(host.OS)
	log.Println(host.Ports)
	log.Println(host.IP)
	log.Println(host.ISP)
	log.Println(host.Hostnames)
	log.Println(host.Organization)
	log.Println(host.Vulnerabilities)
	log.Println(host.ASN)
	log.Println(host.LastUpdate)
	log.Println(host.Data)
	log.Println(host.HostLocation)
}

// HandleMQ : Recv data from the MQ
func (S Shodan) HandleMQ() error {
	return nil
}

// SendMQ : Send data to the MQ
func (S Shodan) SendMQ(data []byte) error {
	return nil
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
		fmt.Println(err)
		return err
	}
	return nil
}

func (S Shodan) WriteDb(result modules.Result, mgoSession *mgo.Session, projectName string) error {
	log.Println("Write to the database.")
	var data ShodanResult
	data = result.Data.(ShodanResult)

	raw := bson.M{projectName + ".results." + result.Module: data}
	database.UpsertRawData(mgoSession, projectName, raw)
	return nil
}

//command: curl -XPOST http://localhost:8080/api/v1/projects/FirstProject/run/shodan
