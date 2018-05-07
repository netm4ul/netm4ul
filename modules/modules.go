package modules

import (
	"net"
	"time"

	"github.com/netm4ul/netm4ul/core/database/models"
)

// Condition defined for dependencies tree
type Condition struct {
	Op     string // "OR", "AND"
	Module string // Module name
}

type Result struct {
	Error     error
	Timestamp time.Time   // Represent the final update timestamp
	Module    string      // Module name
	Data      interface{} // Raw data
}

// Input define the basic target system. Each module can query the database for more information.
type Input struct {
	Domain    string `json:"domain,omitempty"`
	IP        net.IP `json:"ip,omitempty"`
	Port      int16  `json:"port,omitempty"`
	Ressource string `json:"ressource,omitempty"`
}

type Module interface {
	Name() string
	Version() string
	Author() string
	DependsOn() []Condition
	Run([]Input) (Result, error)
	ParseConfig() error
	WriteDb(result Result, db models.Database, projectName string) error
}
