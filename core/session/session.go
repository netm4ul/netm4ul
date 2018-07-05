package session

import (
	"crypto/tls"
	"errors"
	"net"
	"strconv"
	"strings"

	"github.com/netm4ul/netm4ul/core/communication"
	"github.com/netm4ul/netm4ul/core/config"
	"github.com/netm4ul/netm4ul/core/loadbalancing"
	"github.com/netm4ul/netm4ul/modules"
	"github.com/netm4ul/netm4ul/modules/recon/dns"
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
	Nodes          []communication.Node
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
	s.Register(dns.NewDNS())
	s.Register(nmap.NewNmap())
	s.Register(masscan.NewMasscan())
	s.Register(shodan.NewShodan())
}

// GetServerIPPort func
func (s *Session) GetServerIPPort() string {
	return s.Config.Server.IP + ":" + strconv.FormatUint(uint64(s.Config.Server.Port), 10)
}

// GetAPIIPPort fun
func (s *Session) GetAPIIPPort() string {
	return s.Config.Server.IP + ":" + strconv.FormatUint(uint64(s.Config.API.Port), 10)
}
