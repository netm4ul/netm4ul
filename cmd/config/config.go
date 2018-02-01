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
	Port     uint16
	User     string
	Password string
}

// Keys : setup tocken & api keys
type Keys struct {
	Google string
	Shodan string
}

// Server : Master node config
type Server struct {
	User     string
	Password string
	IP       net.IP
	Port     uint16
}

// Module : Basic struct for general module config
type Module struct {
	Enabled bool
}

// Node : Node info
type Node struct {
	Modules []string
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
