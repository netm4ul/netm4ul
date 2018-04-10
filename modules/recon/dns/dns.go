package dns

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/miekg/dns"
	"github.com/netm4ul/netm4ul/cmd/server/database"
	"github.com/netm4ul/netm4ul/modules"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// DnsResult represent the parsed ouput
type DnsResult struct {
	Types    map[string][]string
	resolver string
	err      string
}

// ConfigToml : configuration model (from the toml file)
type ConfigToml struct {
	MaxHops int `toml:"max_hops"`
}

// Dns "class"
type Dns struct {
	// Config : exported config
	Config ConfigToml
}

//NewDns generate a new Dns module (type modules.Module)
func NewDns() modules.Module {
	gob.Register(DnsResult{})
	var d modules.Module
	d = Dns{}
	return d
}

// Name : name getter
func (D Dns) Name() string {
	return "Dns"
}

// Author : Author getter
func (D Dns) Author() string {
	return "Rzbaa"
}

// Version : Version  getter
func (D Dns) Version() string {
	return "0.1"
}

// DependsOn : Generate the dependencies requirement
func (D Dns) DependsOn() []modules.Condition {
	var _ modules.Condition
	return []modules.Condition{}
}

/*
	Usefull command
	curl -XPOST http://localhost:8080/api/v1/projects/FirstProject/run/dns
	check db: Db.projects.find()
	db.projects.remove({})
*/

// Run : Main function of the module
func (D Dns) Run(data []string) (modules.Result, error) {

	// Banner
	fmt.Println("DNS world!")

	/*
		- [ ] DNS types
		- [ ] DNS server IP in config file
	*/

	// Get fqdn of domain
	domain := "edznux.fr"
	fqdnDomain := dns.Fqdn(domain)

	// instanciate DnsResult
	result := DnsResult{}

	// Get DNS IP address from our custom resolv.conf like file
	config, _ := dns.ClientConfigFromFile("modules/recon/dns/resolv.conf")

	// Set dns client parameters
	cli := new(dns.Client)

	// Create DNS request
	request := new(dns.Msg)

	// Set recursion to true
	request.RecursionDesired = true

	// x := make(map[string][]string)
	// x["key"] = append(x["key"], "value")

	// Map Types for DnsResult{} treatment
	result.Types = make(map[string][]string)

	// For all Type in dns library
	for _, index := range dns.TypeToString {

		// Set question with Type flag
		request.SetQuestion(fqdnDomain, dns.StringToType[index])

		// Send request to DNS server
		reply, _, err := cli.Exchange(request, config.Servers[0]+":"+config.Port)

		// Catch error
		if err != nil {
			log.Println(err)
		}

		// Verify DNS flag (error/success)
		if reply.Rcode != dns.RcodeSuccess {
			// If error, put "None" in result field
			result.Types[index] = append(result.Types[index], "None")
			// Log
			log.Println("DNS Request fail for ", index, " type.")
		}

		// Retrieve all replies
		for _, answer := range reply.Answer {
			// Add into result field
			result.Types[index] = append(result.Types[index], answer.String())
			// Log
			log.Println(answer)
		}

	}

	// Return result (DnsResult{}) with timestamp and module name
	return modules.Result{Data: result, Timestamp: time.Now(), Module: D.Name()}, nil
}

// ParseConfig : Load the config from the config folder
func (D Dns) ParseConfig() error {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	configPath := filepath.Join(exPath, "config", "dns.conf")

	if _, err := toml.DecodeFile(configPath, &D.Config); err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(D.Config.MaxHops)
	return nil
}

// WriteDb : Save data
func (D Dns) WriteDb(result modules.Result, mgoSession *mgo.Session, projectName string) error {
	log.Println("Write to the database.")
	var data DnsResult
	data = result.Data.(DnsResult)

	raw := bson.M{projectName + ".results." + result.Module: data}
	database.UpsertRawData(mgoSession, projectName, raw)
	return nil
}
