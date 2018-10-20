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

type Module interface {
	Name() string
	Version() string
	Author() string
	DependsOn() []Condition
	Run(communication.Input, chan communication.Result) (communication.Done, error)
	ParseConfig() error
	WriteDb(result communication.Result, db models.Database, projectName string) error
}

type Report interface {
	Name() string
	Generate(name string) error
}
