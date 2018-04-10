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
	A      string
	AA     string
	CNAME  string
	DNSKEY string
	DS     string
	KEY    string
	MX     string
	NS     string
	PTR    string
	TXT    string
	SRV    string
	SAO    string
	err    string
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

// curl -XPOST http://localhost:8080/api/v1/projects/FirstProject/run/dns
// check db: Db.projects.find()
// Run : Main function of the module
func (D Dns) Run(data []string) (modules.Result, error) {
	fmt.Println("DNS world!")

	/*
		- [ ] A
		- [ ] AA
		- [ ] CNAME
		- [ ] DNSKEY
		- [ ] DS
		- [ ] KEY
		- [ ] MX
		- [ ] NS
		- [ ] PTR
		- [ ] TXT
		- [ ] SRV
		- [ ] SAO
		- [ ] DNS server IP in config file
	*/

	// Get fqdn of domain
	domain := "edznux.fr"
	fqdnDomain := dns.Fqdn(domain)
	// instanciate DnsResult
	result := new(DnsResult)

	// Get DNS IP address from our custom resolv.conf like file
	config, _ := dns.ClientConfigFromFile("modules/recon/dns/resolv.conf")

	// Set dns client parameters
	cli := new(dns.Client)
	// Create DNS request
	request := new(dns.Msg)
	// Set recursion to true
	request.RecursionDesired = true

	// Specify DNS field
	request.SetQuestion(fqdnDomain, dns.TypeA)

	reply, _, err := cli.Exchange(request, config.Servers[0]+":"+config.Port)

	if err != nil {
		log.Println(err)
	}
	if reply.Rcode != dns.RcodeSuccess {
		log.Println(reply.Rcode)
		return modules.Result{Data: DnsResult{err: "Failure DNS"}, Timestamp: time.Now(), Module: D.Name()}, nil
	}
	for _, afield := range reply.Answer {
		if a, ok := afield.(*dns.A); ok {
			result.A = a.String()
		}
	}
	log.Printf("Return %s\n", result.A)
	return modules.Result{Data: result, Timestamp: time.Now(), Module: D.Name()}, nil

	/*
		config, _ := dns.ClientConfigFromFile("modules/recon/dns/resolv.conf")
		c := new(dns.Client)
		m := new(dns.Msg)
		m.SetQuestion(dns.Fqdn("edznux.fr"), dns.TypeA)
		m.RecursionDesired = true
		r, _, err := c.Exchange(m, config.Servers[0]+":"+config.Port)
		if err != nil {
			log.Println(err)
		}
		if r.Rcode != dns.RcodeSuccess {
			log.Println(r.Rcode)
			return modules.Result{Data: DnsResult{Test: "Failure DNS"}, Timestamp: time.Now(), Module: D.Name()}, nil
		}
		for _, afield := range r.Answer {
			if a, ok := afield.(*dns.A); ok {
				fmt.Printf("%s\n", a.String())
			}
		}

		return modules.Result{Data: DnsResult{Test: "Zgeg"}, Timestamp: time.Now(), Module: D.Name()}, nil*/
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
