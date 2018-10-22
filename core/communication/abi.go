package communication

import (
	"net"
	"time"

	"github.com/netm4ul/netm4ul/core/requirements"
)

//Command represents the communication protocol between clients and the master node
type Command struct {
	Name         string                    `json:"name"`
	Options      Input                     `json:"options"`
	Requirements requirements.Requirements `json:"requirements"`
}

// Node : Node info
type Node struct {
	IP           string
	ID           string
	Modules      []string `json:"modules"`
	Project      string   `json:"project"`
	IsAvailable  bool
	Requirements requirements.Requirements
	Conn         net.Conn
}

/*
Result is the structure used by every modules to send their result in "real time" (at the discretion of the module author)
One modules MAY send this structure multiple times.
It should not send this structure after sending the "Done" struct.
*/
type Result struct {
	ModuleName string
	Timestamp  time.Time
	NodeID     string
	Data       interface{}
}

/*
Done represent the data sent by a module when it has finished all operation.
It is normally sent just before exiting.
If the module errored, it MUST send this structure with the Error field set.
*/
type Done struct {
	ModuleName string
	Timestamp  time.Time
	NodeID     string
	Error      error
}

// Input define the basic target system. Each module can query the database for more information.
type Input struct {
	Domain    string `json:"domain,omitempty"`
	IP        net.IP `json:"ip,omitempty"`
	Port      int16  `json:"port,omitempty"`
	Ressource string `json:"ressource,omitempty"`
}
