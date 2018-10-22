package modules

import (
	"github.com/netm4ul/netm4ul/core/communication"
	"github.com/netm4ul/netm4ul/core/database/models"
)

// Condition defined for dependencies tree
type Condition struct {
	Op     string // "OR", "AND"
	Module string // Module name
}

// Module is the minimal interface needed for one module
// The Run function is the "main" function of each module
// ParseConfig will be run in the init system
// The WriteDb is called each time a new result is sent in the "communication.Result chan"
type Module interface {
	Name() string
	Version() string
	Author() string
	DependsOn() []Condition
	Run(communication.Input, chan communication.Result) (communication.Done, error)
	ParseConfig() error
	WriteDb(result communication.Result, db models.Database, projectName string) error
}

// Report is the minimal interface for one reporting module
// The Generate function is the "main" function of the report module
type Report interface {
	Name() string
	Generate(name string) error
}
