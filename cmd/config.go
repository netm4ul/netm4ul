package cmd

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"net"
	"os"
	"path/filepath"
)

type Api struct {
	Port     uint16
	User     string
	Password string
}

type Keys struct {
	Google string
	Shodan string
}

type MQ struct {
	User     string
	Password string
	Ip       net.IP
	Port     uint16
}

type Module struct {
	Enabled bool
}

type Server struct {
	Ip      net.IP
	Modules []string
	Type    string
}

type ConfigToml struct {
	Api     Api
	Keys    Keys
	MQ      MQ
	Servers map[string]Server
	Modules map[string]Module
}

// Config : exported config
var Config ConfigToml

func init() {
	/*
		Get the executable path.
		From there, get the config.
	*/
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
