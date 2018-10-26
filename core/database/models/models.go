package models

import (
	"net"
	"time"

	"github.com/netm4ul/netm4ul/core/config"
)

/*
	See README for the database "schema"
	This file contains some generic informations.
*/

// Hop defines each "hop" from the host (netm4ul client) to the target.
type Hop struct {
	IP  net.IP
	Max float32
	Min float32
	Avg float32
}

// Route defines the route from the host (netm4ul client) to the target
type Route struct {
	Source      string `json:"source,omitempty"`
	Destination string `json:"destination,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// URI defines one ressource from a remote target (webserver), either files or directory
type URI struct {
	Name string `json:"name"`
	Code string `json:"code,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Port defines the basic structure for each port scanned on the target
type Port struct {
	Number   int16  `json:"number,omitempty"`
	Protocol string `json:"protocol,omitempty"`
	Status   string `json:"status,omitempty"` // open, filtered, closed
	Banner   string `json:"banner,omitempty"`
	Type     string `json:"type,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

//IP defines the IP address of a target.
type IP struct {
	Value   string `json:"value,omitempty" gorm:"primary_key"`   // should be net.IP, but can't enforce that in the db...
	Network string `sql:"default:'external'" gorm:"primary_key"` // arbitrary value, default should be "external".

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

//Network represent the network used by one IP.
//By default, every ip should be in the "external" Network
type Network struct {
	Name        string `json:"name"`
	Description string `json:"description"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

//Domain defines the Domain address of a target.
type Domain struct {
	Name string `json:"name,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

//Project is the top level struct for a target. It contains a list of IPs and other metadata.
type Project struct {
	Name        string `json:"name" sql:",unique"`
	Description string `json:"description"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

//User is the default struct for each users.
// The token is a random char stored in the database and needed for all the API calls (except the few non authenticated ones)
type User struct {
	Name     string `json:"name" sql:",unique"`
	Password string `json:"password,omitempty"`
	Token    string `json:"token,omitempty" toml:"token"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

//PortType is a categorie for one ports.
type PortType struct {
	Type        string
	Description string
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Raw defines the base structure for each module data.
// One module must save at least this one inside it's WriteDb function.
// The module should dump all it's result inside the Content field.
type Raw struct {
	Content    string
	ModuleName string
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

//GetPortTypes returns all the port type to insert into the database during the setup. It might be augmented during runtime by the user interaction.
func GetPortTypes() []PortType {
	pts := []PortType{
		{Type: "web", Description: "This port runs a web services (http, https)"},
		{Type: "mail", Description: "This port runs a mail service"},
		{Type: "admin", Description: "This port runs an admin tool, it could be a web page (http), a remote desktop application (rdp), or a remote login (ssh, telnet)"},
		{Type: "gaming", Description: "This port is used by a game"},
		{Type: "voip", Description: "This port is used by a voip service (Teamspeak, Skype, Mumble...)"},
		{Type: "api", Description: "This port serve an API"},
	}
	return pts
}

/*
You can generate a new adapter automatically with the "create". (see `netm4ul create --help` command)
This interface sets the abi for the adapters.

This interface is made to be as database agnostic as possible.
We do not infer on how to store the data, nor how you do "relationship" between items etc...
*/

//Database is the mandatory interface for all custom database adapter
type Database interface {
	// General purpose functions
	Name() string
	SetupDatabase() error
	DeleteDatabase() error
	SetupAuth(username, password, dbname string) error
	Connect(*config.ConfigToml) error

	//Users
	CreateOrUpdateUser(user User) error
	GetUser(username string) (User, error)
	GetUserByToken(token string) (User, error)
	GenerateNewToken(user User) error
	DeleteUser(user User) error

	// Project
	CreateOrUpdateProject(Project) error
	GetProjects() ([]Project, error)
	GetProject(projectName string) (Project, error)
	DeleteProject(project Project) error

	// IP
	CreateOrUpdateIP(projectName string, ip IP) error
	CreateOrUpdateIPs(projectName string, ip []IP) error
	GetIPs(projectName string) ([]IP, error)
	GetIP(projectName string, ip string) (IP, error)
	DeleteIP(ip IP) error

	// Domain
	CreateOrUpdateDomain(projectName string, domain Domain) error
	CreateOrUpdateDomains(projectName string, domain []Domain) error
	GetDomains(projectName string) ([]Domain, error)
	GetDomain(projectName string, domain string) (Domain, error)
	DeleteDomain(projectName string, domain Domain) error

	// Port
	CreateOrUpdatePort(projectName string, ip string, port Port) error
	CreateOrUpdatePorts(projectName string, ip string, port []Port) error
	GetPorts(projectName string, ip string) ([]Port, error)
	GetPort(projectName string, ip string, port string) (Port, error) // TOFIX : includes protocols (tcp, upd...) *optionnal* parameters. Returns CodeAmbiguousRequest if 2 ports of differents type are found.
	DeletePort(projectName string, ip string, port Port) error

	// URI (directory and files)
	CreateOrUpdateURI(projectName string, ip string, port string, dir URI) error
	CreateOrUpdateURIs(projectName string, ip string, port string, dir []URI) error
	GetURIs(projectName string, ip string, port string) ([]URI, error)
	GetURI(projectName string, ip string, port string, dir string) (URI, error)
	DeleteURI(projectName string, ip string, port string, dir URI) error

	// Raw data
	AppendRawData(projectName string, data Raw) error
	GetRaws(projectName string) ([]Raw, error)
	GetRawModule(projectName string, moduleName string) (map[string][]Raw, error)
}
