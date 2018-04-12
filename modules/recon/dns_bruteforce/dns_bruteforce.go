package dnsbf

import (

	"fmt"
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
	codes           string   `toml:"codes"`
	OutputFileName  string   `toml:"OutputFileName"`
	Url             string   `toml:"Url"`
	Username        string   `toml:"Username"`
	Password        string   `toml:"Password"`
	extensions      string   `toml:"extension"`
	UserAgent       string   `toml:"UserAgent"`
	proxy           string   `toml:"proxy"`
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

	s := libgobuster.InitState()

	if D.Config.Threads {
		s.Threads = D.Config.Threads
	}

	if D.Config.Mode {
		s.Mode = D.Config.Mode
	}

	if D.Config.Wordlist {
		s.Wordlist = D.Config.Wordlist
	}

	if D.Config.codes {
		s.codes = D.Config.codes
	}

	if D.Config.OutputFileName {
		s.OutputFileName = D.Config.OutputFileName
	}

	if D.Config.Url {
		s.Url = D.Config.Url
	}

	if D.Config.Username {
		s.Username = D.Config.Username
	}

	if D.Config.Password {
		s.Password = D.Config.Password
	}

	if D.Config.extensions {
		s.extensions = D.Config.extensions
	}

	if D.Config.UserAgent {
		s.UserAgent = D.Config.UserAgent
	}

	if D.Config.proxy {
		s.proxy = D.Config.proxy
	}

	if D.Config.Verbose {
		s.Verbose = D.Config.Verbose
	}

	if D.Config.ShowIPs {
		s.ShowIPs = D.Config.ShowIPs
	}

	if D.Config.ShowCNAME {
		s.ShowCNAME = D.Config.ShowCNAME
	}

	if D.Config.FollowRedirect {
		s.FollowRedirect = D.Config.FollowRedirect
	}

	if D.Config.Quiet {
		s.Quiet = D.Config.Quiet
	}

	if D.Config.Expanded {
		s.Expanded = D.Config.Expanded
	}

	if D.Config.NoStatus {
		s.NoStatus = D.Config.NoStatus
	}

	if D.Config.IncludeLength {
		s.IncludeLength = D.Config.IncludeLength
	}

	if D.Config.UseSlash {
		s.UseSlash = D.Config.UseSlash
	}

	if D.Config.WildcardForced {
		s.WildcardForced = D.Config.WildcardForced
	}

	if D.Config.InsecureSSL {
		s.InsecureSSL = D.Config.InsecureSSL
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

//WriteDb : Save data
func (D *DnsBF) WriteDb(result modules.Result, mgoSession *mgo.Session, projectName string) error {
	log.Println("Write to the database.")
	var data DnsResult // change var ?
	data = result.Data.(DnsResult) // change var ?

	raw := bson.M{projectName + ".results." + result.Module: data}
	database.UpsertRawData(mgoSession, projectName, raw)
	return nil
}