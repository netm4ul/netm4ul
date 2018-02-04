package modules

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
	Run(interface{}) (interface{}, error)
	Parse() (interface{}, error)
	HandleMQ() error
	SendMQ(data []byte) error
	ParseConfig() error
	WriteDb() error
}
