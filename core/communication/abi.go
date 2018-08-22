package communication

import (
	"net"

	"github.com/netm4ul/netm4ul/core/requirements"
	"github.com/netm4ul/netm4ul/modules"
)

//Command represents the communication protocol between clients and the master node
type Command struct {
	Name         string                    `json:"name"`
	Options      modules.Input             `json:"options"`
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
