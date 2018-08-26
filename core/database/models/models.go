package models

import (
	"crypto/rand"
	"fmt"
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
	Source      string `json:"source,omitempty" bson:"Source"`
	Destination string `json:"destination,omitempty" bson:"Destination"`

	CreatedAt time.Time `json:"created_at" bson:"CreatedAt,omitempty"`
	UpdatedAt time.Time `json:"updated_at" bson:"UpdatedAt,omitempty"`
}

// URI defines one ressource from a remote target (webserver), either files or directory
type URI struct {
	Name string `json:"name" bson:"Name"`
	Code string `json:"code,omitempty" bson:"Code,omitempty"`

	CreatedAt time.Time `json:"created_at" bson:"CreatedAt,omitempty"`
	UpdatedAt time.Time `json:"updated_at" bson:"UpdatedAt,omitempty"`
}

// Port defines the basic structure for each port scanned on the target
type Port struct {
	Number   int16  `json:"number,omitempty" bson:"Number"`
	Protocol string `json:"protocol,omitempty" bson:"Protocol"`
	Status   string `json:"status,omitempty" bson:"Status"` // open, filtered, closed
	Banner   string `json:"banner,omitempty" bson:"Banner,omitempty"`
	Type     string `json:"type,omitempty" bson:"Type,omitempty"`

	CreatedAt time.Time `json:"created_at" bson:"CreatedAt,omitempty"`
	UpdatedAt time.Time `json:"updated_at" bson:"UpdatedAt,omitempty"`
}

//IP defines the IP address of a target.
type IP struct {
	Value   string `json:"value,omitempty" bson:"Value" gorm:"primary_key"` // should be net.IP, but can't enforce that in the db...
	Network string `sql:"default:'external'" gorm:"primary_key"`            // arbitrary value, default should be "external".

	CreatedAt time.Time `json:"created_at" bson:"CreatedAt,omitempty"`
	UpdatedAt time.Time `json:"updated_at" bson:"UpdatedAt,omitempty"`
}

//Network represent the network used by one IP.
//By default, every ip should be in the "external" Network
type Network struct {
	Name        string `json:"name"`
	Description string `json:"description"`

	CreatedAt time.Time `json:"created_at" bson:"CreatedAt,omitempty"`
	UpdatedAt time.Time `json:"updated_at" bson:"UpdatedAt,omitempty"`
}

//Domain defines the Domain address of a target.
type Domain struct {
	Name string `json:"name,omitempty" bson:"Name"`

	CreatedAt time.Time `json:"created_at" bson:"CreatedAt,omitempty"`
	UpdatedAt time.Time `json:"updated_at" bson:"UpdatedAt,omitempty"`
}

//Project is the top level struct for a target. It contains a list of IPs and other metadata.
type Project struct {
	Name        string `json:"name" bson:"Name" sql:",unique"`
	Description string `json:"description" bson:"Description,omitempty"`

	CreatedAt time.Time `json:"created_at" bson:"CreatedAt,omitempty"`
	UpdatedAt time.Time `json:"updated_at" bson:"UpdatedAt,omitempty"`
}

type User struct {
	Name     string `json:"name" bson:"Name" sql:",unique"`
	Password string `json:"password,omitempty"`
	Token    string `json:"token,omitempty" toml:"token"`

	CreatedAt time.Time `json:"created_at" bson:"CreatedAt,omitempty"`
	UpdatedAt time.Time `json:"updated_at" bson:"UpdatedAt,omitempty"`
}

type PortType struct {
	Type        string
	Description string
	CreatedAt   time.Time `json:"created_at" bson:"CreatedAt,omitempty"`
	UpdatedAt   time.Time `json:"updated_at" bson:"UpdatedAt,omitempty"`
}

type Raw struct {
	Content    string
	ModuleName string
	CreatedAt  time.Time `json:"created_at" bson:"CreatedAt,omitempty"`
	UpdatedAt  time.Time `json:"updated_at" bson:"UpdatedAt,omitempty"`
}

//GenerateNewToken return a new random token string
func GenerateNewToken() string {
	b := make([]byte, 20)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

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
	GetPort(projectName string, ip string, port string) (Port, error)
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
