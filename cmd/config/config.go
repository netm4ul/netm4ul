package config

import (
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// API : Rest API config
type API struct {
	Port     uint16 `toml:"port"`
	User     string `toml:"user"`
	Password string `toml:"password"`
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
	IP       net.IP `toml:"ip"`
	Port     uint16 `toml:"port"`
}

// Module : Basic struct for general module config
type Module struct {
	Enabled bool `toml:"enabled"`
}

// Node : Node info
type Node struct {
	Modules []string
	Project string
}

// ConfigToml is the global config object
type ConfigToml struct {
	IsServer bool
	API      API
	Keys     Keys
	Server   Server
	Nodes    map[string]Node
	Modules  map[string]Module
}

// Config : exported config
var Config ConfigToml

//	Get the executable path.
//	From there, get the config.
func init() {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}

	exPath := filepath.Dir(ex)
	configPath := filepath.Join(exPath, "netm4ul.conf")

	if _, err := toml.DecodeFile(configPath, &Config); err != nil {
		fmt.Println(err)
		return
	}
}
