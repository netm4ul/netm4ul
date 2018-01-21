package cmd

import(
	"fmt"
	"github.com/BurntSushi/toml"
	"net"
)

type Api struct{
	Port uint16
	User string
	Password string
}

type Keys struct{
	Google string
	Shodan string
}

type MQ struct{
	User string
	Password string
	Ip net.IP
	Port uint16
}

type Module struct{
	enabled bool
}

type Server struct{
	Ip net.IP
	Modules []string
	Type string
}

type ConfigToml struct{
	Api Api
	Keys Keys
	MQ MQ
	Servers Server
	Modules Module
}

// exported config
var Config ConfigToml

func init(){
	if _, err := toml.DecodeFile("../netm4ul.conf", &Config); err != nil{
		fmt.Println(err)
		return
	}
}
