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
	"github.com/netm4ul/netm4ul/cmd/config"
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
func (D *Dns) Run(data []string) (modules.Result, error) {

	// Banner
	fmt.Println("DNS world!")

	// Generate config file
	D.ParseConfig()

	// Get fqdn of domain
	domain := "edznux.fr"
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
		config, _ := dns.ClientConfigFromReader(r)
		resolverConfig[i] = config
	}

	// Map Types for DnsResult{} treatment
	result.Types = make(map[string][]string)

	// For all Type in dns library

	if D.Config.RandomDns == true {
		// random DNS resolver
		for _, index := range dns.TypeToString {
			config := resolverConfig[rand.Intn(len(resolverConfig))]
			requestRoutine(index, result, fqdn, config)
		}
	} else {
		// Normal iteration of DNS resolvers
		i := 0
		for _, index := range dns.TypeToString {
			// Get config
			if i == len(resolverConfig) {
				i = 0
			}
			config := resolverConfig[i]
			requestRoutine(index, result, fqdn, config)
			i++
		}
	}

	// Return result (DnsResult{}) with timestamp and module name
	return modules.Result{Data: result, Timestamp: time.Now(), Module: D.Name()}, nil
}

// Forge and send DNS request for index type
func requestRoutine(index string, result DnsResult, fqdn string, config *dns.ClientConfig) {
	// Set dns client parameters
	cli := new(dns.Client)

	// Create DNS request
	request := new(dns.Msg)

	// Set recursion to true
	request.RecursionDesired = true

	// Set question with Type flag
	request.SetQuestion(fqdn, dns.StringToType[index])

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
		log.Println("DNS Request fail. No", index, "type.")
	}

	// Retrieve all replies
	for _, answer := range reply.Answer {

		// Add into result field
		dnsParser(answer, index, result)
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
func (D *Dns) WriteDb(result modules.Result, mgoSession *mgo.Session, projectName string) error {
	log.Println("Write to the database.")
	var data DnsResult
	data = result.Data.(DnsResult)

	raw := bson.M{projectName + ".results." + result.Module: data}
	database.UpsertRawData(mgoSession, projectName, raw)
	return nil
}

// Shit happens
// DNS parser
func dnsParser(answer dns.RR, index string, result DnsResult) {
	switch t := answer.(type) {
	case *dns.A:
		result.Types[index] = append(result.Types[index], t.A.String())
	case *dns.AAAA:
		result.Types[index] = append(result.Types[index], t.AAAA.String())
	case *dns.AFSDB:
		result.Types[index] = append(result.Types[index], t.Hostname)
	case *dns.ANY:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.AVC:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.CAA:
		result.Types[index] = append(result.Types[index], t.Value)
	case *dns.CDNSKEY:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.CDS:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.CERT:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.CNAME:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.CSYNC:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.DHCID:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.DLV:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.DNAME:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.DNSKEY:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.DS:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.EID:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.EUI48:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.EUI64:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.GID:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.GPOS:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.HINFO:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.HIP:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.KEY:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.KX:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.L32:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.L64:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.LOC:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.LP:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.MB:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.MD:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.MF:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.MG:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.MINFO:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.MR:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.MX:
		result.Types[index] = append(result.Types[index], t.Mx)
	case *dns.NAPTR:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.NID:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.NIMLOC:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.NINFO:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.NS:
		result.Types[index] = append(result.Types[index], t.Ns)
	case *dns.NSAPPTR:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.NSEC:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.NSEC3:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.NSEC3PARAM:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.OPENPGPKEY:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.OPT:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.PTR:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.PX:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.RKEY:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.RP:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.RRSIG:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.RT:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.SIG:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.SMIMEA:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.SOA:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.SPF:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.SRV:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.SSHFP:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.TA:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.TALINK:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.TKEY:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.TLSA:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.TSIG:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.TXT:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.UID:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.UINFO:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.URI:
		result.Types[index] = append(result.Types[index], t.String())
	case *dns.X25:
		result.Types[index] = append(result.Types[index], t.String())
	default:
		result.Types[index] = append(result.Types[index], "DnsParserError")
	}
}
