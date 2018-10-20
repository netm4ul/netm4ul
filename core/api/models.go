package api

import (
	"github.com/gorilla/mux"
	"github.com/netm4ul/netm4ul/core/communication"
	"github.com/netm4ul/netm4ul/core/database/models"
	"github.com/netm4ul/netm4ul/core/server"
	"github.com/netm4ul/netm4ul/core/session"
)

// Result is the standard response format
type Result struct {
	Status   string      `json:"status"`
	Code     Code        `json:"code"`
	Message  string      `json:"message,omitempty"`
	Data     interface{} `json:"data,omitempty"`
	HTTPCode int         `json:"-"` //remove HTTPCode from the json response
}

//API is the constructor for this package
type API struct {
	// Session defines the global session for the API.
	Session *session.Session
	Server  *server.Server
	db      models.Database
	Prefix  string
	Router  *mux.Router
	IPPort  string
	Version string
}

//Info provides general purpose information for this API
type Info struct {
	Port     uint16 `json:"port,omitempty"`
	Versions string `json:"versions"`
}

//Metadata of the current system (node, api, database)
type Metadata struct {
	Nodes []communication.Node `json:"nodes"`
	Info  Info                 `json:"api"`
}

// CreateAPI : Initialise the infinite server loop on the master node
func CreateAPI(s *session.Session, server *server.Server) *API {
	api := API{
		Session: s,
		Server:  server,
		db:      server.Db,
	}

	return &api
}
