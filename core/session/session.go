package session

import (
	"strings"

	"github.com/netm4ul/netm4ul/modules"
<<<<<<< HEAD:cmd/session/session.go
	"github.com/netm4ul/netm4ul/modules/recon/dns"
=======
	"github.com/netm4ul/netm4ul/modules/recon/nmap"
>>>>>>> develop:core/session/session.go
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
<<<<<<< HEAD:cmd/session/session.go
	p.Register(dns.NewDns())
=======
	p.Register(nmap.NewNmap())
>>>>>>> develop:core/session/session.go
}
