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

/*	A          []string
	AAAA       []string
	AFSDB      []string
	ANY        []string
	AVC        []string
	CAA        []string
	CDNSKEY    []string
	CDS        []string
	CERT       []string
	CNAME      []string
	CSYNC      []string
	DHCID      []string
	DLV        []string
	DNAME      []string
	DNSKEY     []string
	DS         []string
	EID        []string
	EUI48      []string
	EUI64      []string
	GID        []string
	GPOS       []string
	HINFO      []string
	HIP        []string
	KEY        []string
	KX         []string
	L32        []string
	L64        []string
	LOC        []string
	LP         []string
	MB         []string
	MD         []string
	MF         []string
	MG         []string
	MINFO      []string
	MR         []string
	MX         []string
	NAPTR      []string
	NID        []string
	NIMLOC     []string
	NINFO      []string
	NS         []string
	NSAPPTR    []string
	NSEC       []string
	NSEC3      []string
	NSEC3PARAM []string
	OPENPGPKEY []string
	OPT        []string
	PTR        []string
	PX         []string
	RKEY       []string
	RP         []string
	RRSIG      []string
	RT         []string
	SIG        []string
	SMIMEA     []string
	SOA        []string
	SPF        []string
	SRV        []string
	SSHFP      []string
	TA         []string
	TALINK     []string
	TKEY       []string
	TLSA       []string
	TSIG       []string
	TXT        []string
	UID        []string
	UINFO      []string
	URI        []string
	X25        []string*/
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
curl -XPOST http://localhost:8080/api/v1/projects/FirstProject/run/dns
check db: Db.projects.find()
db.projects.remove({})
*/
// Run : Main function of the module
func (D Dns) Run(data []string) (modules.Result, error) {
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

	for _, index := range dns.TypeToString {
		request.SetQuestion(fqdnDomain, dns.StringToType[index])

		//request.SetQuestion(fqdnDomain, dns.StringToType[index])
		reply, _, err := cli.Exchange(request, config.Servers[0]+":"+config.Port)

		if err != nil {
			log.Println(err)
		}
		if reply.Rcode != dns.RcodeSuccess {
			result.Types = make(map[string][]string)
			result.Types[index] = append(result.Types[index], "None")
			log.Println("DNS Request fail for ", index, " type.")
		}

		for _, answer := range reply.Answer {
			log.Println(answer)
		}
	}
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
