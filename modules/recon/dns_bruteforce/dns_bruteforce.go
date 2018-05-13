package dnsbf

import (

	"fmt"
	"time"
	"log"

	"encoding/gob"
	"errors"
	log "github.com/sirupsen/logrus"

	//"github.com/netm4ul/netm4ul/core/database/models"
	"github.com/netm4ul/netm4ul/modules"

	"github.com/OJ/gobuster/libgobuster"
	"github.com/BurntSushi/toml"

)

// DnsResult represent the parsed ouput
type DnsResult struct {

}

//ConfigToml : configuration model (from the toml file)
type ConfigToml struct {
	Threads         int      `toml:"Threads"`
	Mode            string   `toml:"Mode"`
	Wordlist        string   `toml:"Wordlist"`
	codes           string   `toml:"Codes"`
	OutputFileName  string   `toml:"OutputFileName"`
	Url             string   `toml:"Url"`
	Username        string   `toml:"Username"`
	Password        string   `toml:"Password"`
	extensions      string   `toml:"Extension"`
	UserAgent       string   `toml:"UserAgent"`
	proxy           string   `toml:"Proxy"`
	Verbose         bool     `toml:"Verbose"`
	ShowIPs         bool     `toml:"ShowIPs"`
	ShowCNAME       bool     `toml:"ShowCNAME"`
	FollowRedirect  bool     `toml:"FollowRedirect"`
	Quiet           bool     `toml:"Quiet"`
	Expanded        bool     `toml:"Expanded"`
	NoStatus        bool     `toml:"NoStatus"`
	IncludeLength   bool     `toml:"IncludeLength"`
	UseSlash        bool     `toml:"UseSlash"`
	WildcardForced  bool     `toml:"WildcardForced"`
	InsecureSSL     bool     `toml:"InsecureSSL"`
}

// DnsBF "class"
type DnsBF struct {
	Config ConfigToml
	State *libgobuster.State
}

//NewDns generate a new Dns module (type modules.Module)
func NewDnsBF() modules.Module {
	gob.Register(DnsResult{}) // change var ?
	var d modules.Module
	d = DnsBF{}
	return d
}

// Name : name getter
func (D *DnsBF) Name() string {
	return "DnsBF"
}

// Author : Author getter
func (D *DnsBF) Author() string {
	return "Skawak"
}

// Version : Version  getter
func (D *DnsBF) Version() string {
	return "0.1"
}

// DependsOn : Generate the dependencies requirement
func (D *DnsBF) DependsOn() []modules.Condition {
	var _ modules.Condition
	return []modules.Condition{}
}

// Run : Main function of the module
func (D *DnsBF) Run(data []string) (modules.Result, error) {

	// Banner
	fmt.Println("DNS BruteForce")

	// Let's go
	D.ParseConfig()

	D.State := libgobuster.InitState()

	D.ParseArgGlobal()

	switch D.Config.Mode {
	case "dir":
		D.ParseArgDIR()
		res = D.Execute()
	case "dns":
		D.ParseArgDNS()
		res = D.Execute()
	default:
		log.Println("Error, this mode doesn't exist")
	}

	fmt.Println("Result")
	fmt.Println(res)

	//return modules.Result{Data: D.Result, Timestamp: time.Now(), Module: D.Name()}//, err

}

func (D *DnsBF) Execute() {
	if err := libgobuster.ValidateState(D.State, extensions, codes, proxy); err.ErrorOrNil() != nil {
		fmt.Printf("%s\n", err.Error())
		//return nil
	} else {
		libgobuster.Process(D.State)
	}
}

// ParseArgGlobal : verify global argument
func (D *DnsBF) ParseArgGlobal() {
	// Number of  threads
	if D.Config.Threads {
		D.State.Threads = D.Config.Threads
	}

	// Wordlist to use
	if D.Config.Wordlist {
		D.State.Wordlist = D.Config.Wordlist
	}

	// Output file for results
	if D.Config.OutputFileName {
		D.State.OutputFileName = D.Config.OutputFileName
	}

	// Target URL or domain
	if D.Config.Url {
		D.State.Url = D.Config.Url
	}

	// Verbose output (errors)
	if D.Config.Verbose {
		D.State.Verbose = D.Config.Verbose
	}

	// Follow redirects
	if D.Config.FollowRedirect {
		D.State.FollowRedirect = D.Config.FollowRedirect
	}

	// Don't print the original banner
	if D.Config.Quiet {
		D.State.Quiet = D.Config.Quiet
	}

	// Expanded mode, print full URLs
	if D.Config.Expanded {
		D.State.Expanded = D.Config.Expanded
	}

	// Don't print status codes
	if D.Config.NoStatus {
		D.State.NoStatus = D.Config.NoStatus
	}

	// Force continued operation when wildcard found
	if D.Config.WildcardForced {
		D.State.WildcardForced = D.Config.WildcardForced
	}

	// Skip SSL certificate verification
	if D.Config.InsecureSSL {
		D.State.InsecureSSL = D.Config.InsecureSSL
	}
}

// ParseArgDIR : verify dir mode argument
func (D *DnsBF) ParseArgDIR() {

	D.State.Mode = "dir"

	// Positive status codes (dir mode only)
	if D.Config.codes {
		D.State.codes = D.Config.codes
	}

	// Username for basic Auth (dir mode only)
	if D.Config.Username {
		D.State.Username = D.Config.Username
	}

	// Password for basic Auth (dir mode only)
	if D.Config.Password {
		D.State.Password = D.Config.Password
	}

	// File extention(s) to search for (dir mode only)
	if D.Config.extensions {
		D.State.extensions = D.Config.extensions
	}

	// Set the User-Agent (dir mode only)
	if D.Config.UserAgent {
		D.State.UserAgent = D.Config.UserAgent
	}

	// Proxy use for requests : http(s)://host:port (dir mode only)
	if D.Config.proxy {
		D.State.proxy = D.Config.proxy
	}

	// Include the length of the body in the output (dir mode only)
	if D.Config.IncludeLength {
		D.State.IncludeLength = D.Config.IncludeLength
	}

	// Appand a forward-slash to each directory request (dir mode only)
	if D.Config.UseSlash {
		D.State.UseSlash = D.Config.UseSlash
	}
}

// ParseArgDNS : verify dns mode argument
func (D *DnsBF) ParseArgDNS() {

	D.State.Mode = "dns"

	// Show IP adresses (dns mode only)
	if D.Config.ShowIPs {
		D.State.ShowIPs = D.Config.ShowIPs
	}

	// Show CNAME records (dns mode only)
	if D.Config.ShowCNAME {
		D.State.ShowCNAME = D.Config.ShowCNAME
	}
}

// ParseConfig : Load the config from the config folder
func (D *DnsBF) ParseConfig() error {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	configPath := filepath.Join(exPath, "config", "dnsbf.conf")

	if _, err := toml.DecodeFile(configPath, &D.Config); err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

//WriteDb : Save data
func (D *DnsBF) WriteDb(result modules.Result, mgoSession *mgo.Session, projectName string) error {
	log.Println("Write to the database.")
	var data DnsResult // change var ?
	data = result.Data.(DnsResult) // change var ?

	raw := bson.M{projectName + ".results." + result.Module: data}
	database.UpsertRawData(mgoSession, projectName, raw)
	return nil
}