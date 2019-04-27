package session

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/netm4ul/netm4ul/modules/recon/certificatetransparency"
	"net"
	"strings"

	"github.com/netm4ul/netm4ul/core/communication"
	"github.com/netm4ul/netm4ul/core/config"
	"github.com/netm4ul/netm4ul/core/loadbalancing"
	"github.com/netm4ul/netm4ul/modules"
	"github.com/netm4ul/netm4ul/modules/recon/dns"
	"github.com/netm4ul/netm4ul/modules/recon/dnsbruteforce"
	"github.com/netm4ul/netm4ul/modules/recon/masscan"
	"github.com/netm4ul/netm4ul/modules/recon/nmap"
	"github.com/netm4ul/netm4ul/modules/recon/shodan"
	"github.com/netm4ul/netm4ul/modules/recon/traceroute"
)

// Connector type, to handle either use of TLS or not
type Connector struct {
	TLSConn *tls.Conn
	Conn    net.Conn
}

// Session type :
type Session struct {
	ModulesEnabled map[string]modules.Module
	Modules        map[string]modules.Module
	Config         config.ConfigToml
	Connector      Connector
	Algo           loadbalancing.Algorithm
	Nodes          map[string]communication.Node
	IsServer       bool
	IsClient       bool
	ConfigPath     string
	Verbose        bool
}

// NewSession func :
func NewSession(c config.ConfigToml) (*Session, error) {
	var err error
	s := Session{
		Modules:        make(map[string]modules.Module, 0),
		ModulesEnabled: make(map[string]modules.Module, 0),
	}
	// populate all modules
	s.Config = c
	s.loadModule()

	s.Algo, err = loadbalancing.NewAlgo(c.Algorithm.Name)

	if err != nil {
		return nil, errors.New("Could not create the session : " + err.Error())
	}
	return &s, nil
}

// Register func :
func (s *Session) Register(m modules.Module) {
	moduleName := strings.ToLower(m.Name())
	s.Modules[moduleName] = m

	if s.Config.Modules[moduleName].Enabled {
		s.ModulesEnabled[moduleName] = m
	}
}

// loadModule func
func (s *Session) loadModule() {
	s.Register(traceroute.NewTraceroute())
	s.Register(certificatetransparency.Newcertificatetransparency())
	s.Register(dns.NewDNS())
	s.Register(nmap.NewNmap())
	s.Register(dnsbruteforce.NewDnsbruteforce())
	s.Register(masscan.NewMasscan())
	s.Register(shodan.NewShodan())
}

// GetServerIPPort func
func (s *Session) GetServerIPPort() string {
	ipport := net.TCPAddr{IP: net.ParseIP(s.Config.Server.IP), Port: int(s.Config.Server.Port)}
	return ipport.String()
}

// GetAPIIPPort func
func (s *Session) GetAPIIPPort() string {
	ipport := net.TCPAddr{IP: net.ParseIP(s.Config.API.IP), Port: int(s.Config.API.Port)}
	return ipport.String()
}

// GetModulesList return a string of all modules (listed and enabled)
// display them in 2 lines
func (s *Session) GetModulesList() string {
	mod := "["
	for _, m := range s.Modules {
		mod += " " + m.Name()
	}
	mod += " ]"

	modEnabled := "["
	for _, m := range s.ModulesEnabled {
		modEnabled += " " + m.Name()
	}
	modEnabled += "]"
	return fmt.Sprintf("Modules : [%s]\n Modules enabled : [%s]", mod, modEnabled)
}
