package modules

import (
	"time"

	mgo "gopkg.in/mgo.v2"
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

type Module interface {
	Name() string
	Version() string
	Author() string
	DependsOn() []Condition
	Run([]string) (Result, error)
	ParseConfig() error
	WriteDb(Result, *mgo.Session, string) error
}
