package dns_bf

import (

	"fmt"
	"github.com/OJ/gobuster/libgobuster"

)

//NewDns generate a new Dns module (type modules.Module)
func NewDns_BF() modules.Module {
	gob.Register(DnsResult{}) // change var ?
	var d modules.Module
	d = Dns_bf{}
	return d
}

// Name : name getter
func (D Dns_bf) Name() string {
	return "Dns_BF"
}

// Author : Author getter
func (D Dns_bf) Author() string {
	return "Skawak"
}

// Version : Version  getter
func (D Dns_bf) Version() string {
	return "0.1"
}

// DependsOn : Generate the dependencies requirement
func (D Dns_bf) DependsOn() []modules.Condition {
	var _ modules.Condition
	return []modules.Condition{}
}

// Run : Main function of the module
func (D Dns_bf) Run(data []string) (modules.Result, error) {

	// Let's go

}

// ParseConfig : Load the config from the config folder
func (D Dns_bf) ParseConfig() error {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	configPath := filepath.Join(exPath, "config", "dns_bf.conf")

	if _, err := toml.DecodeFile(configPath, &D.Config); err != nil {
		fmt.Println(err)
		return err
	}
	return nil

WriteDb : Save data
func (D Dns_bf) WriteDb(result modules.Result, mgoSession *mgo.Session, projectName string) error {
	log.Println("Write to the database.")
	var data DnsResult // change var ?
	data = result.Data.(DnsResult) // change var ?

	raw := bson.M{projectName + ".results." + result.Module: data}
	database.UpsertRawData(mgoSession, projectName, raw)
	return nil
}