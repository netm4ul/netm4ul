package dns

import (
	"encoding/gob"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/miekg/dns"
	"github.com/netm4ul/netm4ul/core/config"
	"github.com/netm4ul/netm4ul/core/database/models"
	"github.com/netm4ul/netm4ul/modules"
)

// DnsResult represent the parsed ouput
type DnsResult struct {
	Types map[string][]string
}

// ConfigToml : configuration model (from the toml file)
type ConfigToml struct {
	RandomDns bool `toml:"randomDNS"`
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
	d = &Dns{}
	return d
}

// Name : name getter
func (D *Dns) Name() string {
	return "Dns"
}

// Author : Author getter
func (D *Dns) Author() string {
	return "Rzbaa"
}

// Version : Version  getter
func (D *Dns) Version() string {
	return "0.1"
}

// DependsOn : Generate the dependencies requirement
func (D *Dns) DependsOn() []modules.Condition {
	var _ modules.Condition
	return []modules.Condition{}
}

/*
	Usefull command
	curl -XPOST http://localhost:8080/api/v1/projects/FirstProject/run/dns
	check db: Db.projects.find()
	remove all data: db.projects.remove({})
*/

// Run : Main function of the module
func (D *Dns) Run(inputs []modules.Input) (modules.Result, error) {

	// Banner
	fmt.Println("DNS world!")

	// Generate config file
	err := D.ParseConfig()
	if err != nil {
		log.Println(err)
	}
	// Get fqdn of domain
	var domain string

	for _, input := range inputs {
		//TODO get each domain name
		if input.Domain != "" {
			domain = input.Domain
			break
		}
	}
	fqdn := dns.Fqdn(domain)

	// instanciate DnsResult
	result := DnsResult{}

	// Get DNS IP address from our config file
	resolverConfig := make(map[int]*dns.ClientConfig)

	resolverList := strings.Split(config.Config.DNS.Resolvers, ",")
	for i := 0; i < len(resolverList); i++ {
		// DNS gog library need resolv.conf entry (nameserver server_ip)
		dnsEntry := "nameserver " + resolverList[i]
		r := strings.NewReader(dnsEntry)
		config, err := dns.ClientConfigFromReader(r)
		if err != nil {
			log.Println(err)
			break
		}
		resolverConfig[i] = config
	}

	// Map Types for DnsResult{} treatment
	result.Types = make(map[string][]string)

	// For all Type in dns library

	if D.Config.RandomDns {
		// random DNS resolver
		for _, dnsType := range dns.TypeToString {
			config := resolverConfig[rand.Intn(len(resolverConfig))]
			requestRoutine(dnsType, result, fqdn, config)
		}
	} else {
		// Normal iteration of DNS resolvers
		i := 0
		for _, dnsType := range dns.TypeToString {
			// Get config
			if i == len(resolverConfig) {
				i = 0
			}
			config := resolverConfig[i]
			requestRoutine(dnsType, result, fqdn, config)
			i++
		}
	}

	// Return result (DnsResult{}) with timestamp and module name
	return modules.Result{Data: result, Timestamp: time.Now(), Module: D.Name()}, nil
}

// Forge and send DNS request for dnsType type
func requestRoutine(dnsType string, result DnsResult, fqdn string, config *dns.ClientConfig) {
	// Set dns client parameters
	cli := new(dns.Client)

	// Create DNS request
	request := new(dns.Msg)

	// Set recursion to true
	request.RecursionDesired = true

	// Set question with Type flag
	request.SetQuestion(fqdn, dns.StringToType[dnsType])

	// Send request to DNS server
	reply, _, err := cli.Exchange(request, config.Servers[0]+":"+config.Port)

	// Catch error
	if err != nil {
		log.Println(err)
	}

	// Verify DNS flag (error/success)
	if reply.Rcode != dns.RcodeSuccess {
		// If error, put "None" in result field
		result.Types[dnsType] = append(result.Types[dnsType], "None")
		// Log
		log.Println("DNS Request fail. No", dnsType, "type.")
	}

	// Retrieve all replies
	for _, answer := range reply.Answer {

		// Add into result field
		dnsParser(answer, dnsType, result)
	}
}

// ParseConfig : Load the config from the config folder
func (D *Dns) ParseConfig() error {
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
	// log.Println(D.Config.RandomDNS)
	return nil
}

// WriteDb : Save data
func (D *Dns) WriteDb(result modules.Result, db models.Database, projectName string) error {
	log.Println("Write to the database.")
	// var data DnsResult
	// data = result.Data.(DnsResult)

	// raw := bson.M{projectName + ".results." + result.Module: data}
	// database.UpsertRawData(mgoSession, projectName, raw)
	return nil
}

// Shit happens
// DNS parser
func dnsParser(answer dns.RR, dnsType string, result DnsResult) {
	switch t := answer.(type) {
	case *dns.A:
		result.Types[dnsType] = append(result.Types[dnsType], t.A.String())
	case *dns.AAAA:
		result.Types[dnsType] = append(result.Types[dnsType], t.AAAA.String())
	case *dns.AFSDB:
		result.Types[dnsType] = append(result.Types[dnsType], t.Hostname)
	case *dns.ANY:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.AVC:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.CAA:
		result.Types[dnsType] = append(result.Types[dnsType], t.Value)
	case *dns.CDNSKEY:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.CDS:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.CERT:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.CNAME:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.CSYNC:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.DHCID:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.DLV:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.DNAME:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.DNSKEY:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.DS:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.EID:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.EUI48:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.EUI64:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.GID:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.GPOS:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.HINFO:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.HIP:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.KEY:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.KX:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.L32:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.L64:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.LOC:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.LP:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.MB:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.MD:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.MF:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.MG:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.MINFO:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.MR:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.MX:
		result.Types[dnsType] = append(result.Types[dnsType], t.Mx)
	case *dns.NAPTR:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.NID:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.NIMLOC:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.NINFO:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.NS:
		result.Types[dnsType] = append(result.Types[dnsType], t.Ns)
	case *dns.NSAPPTR:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.NSEC:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.NSEC3:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.NSEC3PARAM:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.OPENPGPKEY:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.OPT:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.PTR:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.PX:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.RKEY:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.RP:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.RRSIG:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.RT:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.SIG:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.SMIMEA:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.SOA:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.SPF:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.SRV:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.SSHFP:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.TA:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.TALINK:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.TKEY:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.TLSA:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.TSIG:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.TXT:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.UID:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.UINFO:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.URI:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	case *dns.X25:
		result.Types[dnsType] = append(result.Types[dnsType], t.String())
	default:
		result.Types[dnsType] = append(result.Types[dnsType], "DnsParserError")
	}
}
