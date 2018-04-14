package session

import (
	"strings"

	"github.com/netm4ul/netm4ul/modules"
	"github.com/netm4ul/netm4ul/modules/recon/dns"
	"github.com/netm4ul/netm4ul/modules/recon/nmap"
	"github.com/netm4ul/netm4ul/modules/recon/shodan"
	"github.com/netm4ul/netm4ul/modules/recon/traceroute"
)

type Session struct {
	Modules map[string]modules.Module
}

func NewSession() *Session {
	p := Session{
		Modules: make(map[string]modules.Module, 0),
	}
	// populate all modules
	p.loadModule()
	return &p
}

func (p *Session) Register(m modules.Module) {
	p.Modules[strings.ToLower(m.Name())] = m
}

func (p *Session) loadModule() {
	p.Register(traceroute.NewTraceroute())
	p.Register(shodan.NewShodan())
	p.Register(dns.NewDns())
	p.Register(nmap.NewNmap())
}
