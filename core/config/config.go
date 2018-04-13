package config

import (
	"log"
	"os"
	"path/filepath"

	"crypto/tls"
	"github.com/BurntSushi/toml"
)

// API : Rest API config
type API struct {
	Port     uint16 `toml:"port"`
	User     string `toml:"user"`
	Password string `toml:"password"`
}

// DNS : Setup DNS resolver IP
type DNS struct {
	Resolvers string `toml:"resolvers"`
}

// Keys : setup tocken & api keys
type Keys struct {
	Google string `toml:"google"`
	Shodan string `toml:"shodan"`
}

// Server : Master node config
type Server struct {
	User     string `toml:"user"`
	Password string `toml:"password"`
	IP       string `toml:"ip"`
	Port     uint16 `toml:"port"`
}

// Database : Mongodb config
type Database struct {
	User     string `toml:"user"`
	Password string `toml:"password"`
	IP       string `toml:"ip"`
	Port     uint16 `toml:"port"`
}

// Module : Basic struct for general module config
type Module struct {
	Enabled bool `toml:"enabled" json:"enabled"`
}

// Versions : Store the version
type Versions struct {
	Api    string `toml:"api" json:"api"`
	Server string `toml:"server" json:"server"`
	Client string `toml:"client" json:"client"`
}

// Node : Node info
type Node struct {
	Modules []string `json:"modules"`
	Project string   `json:"project"`
}

type Project struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ConfigToml is the global config object
type ConfigToml struct {
	Project    Project
	Versions   Versions
	Verbose    bool
	NoColors   bool
	ConfigPath string
	Mode       string
	IsServer   bool
	IsClient   bool
	Targets    []string
	API        API
	DNS        DNS
	Keys       Keys
	Server     Server
	Database   Database
	Nodes      map[string]Node
	Modules    map[string]Module
	TLSParams  *tls.Config
}

// Config : exported config
var Config ConfigToml

// LoadConfig load the configuration file !
func LoadConfig(file string) {
	var configPath string

	if file == "" {
		dir, err := os.Getwd()

		if err != nil {
			log.Fatal(err)
		}

		configPath = filepath.Join(dir, "netm4ul.conf")
	} else {
		configPath = file
	}

	if _, err := toml.DecodeFile(configPath, &Config); err != nil {
		log.Fatalln(err)
	}
}
