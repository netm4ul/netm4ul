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
	Threads         int      `toml:"threads"`
	Mode            string   `toml:"mode"`
	Wordlist        string   `toml:"wordlist"`
	codes           string   `toml:"codes"`
	OutputFileName  string   `toml:"output"`
	Url             string   `toml:"url"`
	Username        string   `toml:"username"`
	Password        string   `toml:"password"`
	extensions      string   `toml:"extension"`
	UserAgent       string   `toml:"useragent"`
	proxy           string   `toml:"proxy"`
	Verbose         bool     `toml:"verbose"`
	ShowIPs         bool     `toml:"showips"`
	ShowCNAME       bool     `toml:"showcname"`
	FollowRedirect  bool     `toml:"followredirect"`
	Quiet           bool     `toml:"quiet"`
	Expanded        bool     `toml:"expanded"`
	NoStatus        bool     `toml:"nostatus"`
	IncludeLength   bool     `toml:"includelength"`
	UseSlash        bool     `toml:"useslash"`
	WildcardForced  bool     `toml:"wildcardforced"`
	InsecureSSL     bool     `toml:"insecuressl"`
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