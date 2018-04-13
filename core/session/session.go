package session

import (
	"strconv"
	"strings"

	"github.com/netm4ul/netm4ul/core/config"
	"github.com/netm4ul/netm4ul/modules"
	"github.com/netm4ul/netm4ul/modules/recon/traceroute"
	mgo "gopkg.in/mgo.v2"
)

type Session struct {
	Modules      map[string]modules.Module
	Config       config.ConfigToml
	ConnectionDB *mgo.Session
}

func NewSession(c config.ConfigToml) *Session {
	s := Session{
		Modules: make(map[string]modules.Module, 0),
	}
	// populate all modules
	s.Config = c
	s.loadModule()
	return &s
}

func (s *Session) Register(m modules.Module) {
	s.Modules[strings.ToLower(m.Name())] = m
}

func (s *Session) loadModule() {
	s.Register(traceroute.NewTraceroute())
}

func (s *Session) GetServerIPPort() string {
	return s.Config.Server.IP + ":" + strconv.FormatUint(uint64(s.Config.Server.Port), 10)
}

func (s *Session) GetAPIIPPort() string {
	return s.Config.Server.IP + ":" + strconv.FormatUint(uint64(s.Config.API.Port), 10)
}
