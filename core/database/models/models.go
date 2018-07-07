package models

import (
	"crypto/rand"
	"fmt"
	"net"

	"github.com/netm4ul/netm4ul/core/config"
)

/*
	See README for the database "schema"
	This file contains some generic informations.
*/

// Hop defines each "hop" from the host (netm4ul client) to the target.
type Hop struct {
	ID  string `json:"-" bson:"_id,omitempty"`
	IP  net.IP
	Max float32
	Min float32
	Avg float32
}

// Route defines the route from the host (netm4ul client) to the target
type Route struct {
	ID          string `json:"-" bson:"_id,omitempty"`
	Source      string `json:"source,omitempty" bson:"Source"`
	Destination string `json:"destination,omitempty" bson:"Destination"`
	Hops        []Hop  `json:"hops,omitempty" bson:"Hops,omitempty"`

	CreatedAt int64 `json:"created_at" bson:"CreatedAt,omitempty"`
	UpdatedAt int64 `json:"updated_at" bson:"UpdatedAt,omitempty"`
}

// URI defines one ressource from a remote target (webserver), either files or directory
type URI struct {
	ID   string `json:"-" bson:"_id,omitempty"`
	Name string `json:"name" bson:"Name"`
	Code string `json:"code,omitempty" bson:"Code,omitempty"`

	CreatedAt int64 `json:"created_at" bson:"CreatedAt,omitempty"`
	UpdatedAt int64 `json:"updated_at" bson:"UpdatedAt,omitempty"`
}

// Port defines the basic structure for each port scanned on the target
type Port struct {
	ID       string `json:"-" bson:"_id,omitempty"`
	Number   int16  `json:"number,omitempty" bson:"Number"`
	Protocol string `json:"protocol,omitempty" bson:"Protocol"`
	Status   string `json:"status,omitempty" bson:"Status"` // open, filtered, closed
	Banner   string `json:"banner,omitempty" bson:"Banner,omitempty"`
	Type     string `json:"type,omitempty" bson:"Type,omitempty"`
	URIs     []URI  `json:"uris,omitempty" bson:"uris,omitempty"`

	CreatedAt int64 `json:"created_at" bson:"CreatedAt,omitempty"`
	UpdatedAt int64 `json:"updated_at" bson:"UpdatedAt,omitempty"`
}

//IP defines the IP address of a target.
type IP struct {
	ID    string `json:"-" bson:"_id,omitempty"`
	Value string `json:"value,omitempty" bson:"Value"` // should be net.IP, but can't enforce that in the db...
	Ports []Port `json:"ports,omitempty" bson:"Ports,omitempty"`

	CreatedAt int64 `json:"created_at" bson:"CreatedAt,omitempty"`
	UpdatedAt int64 `json:"updated_at" bson:"UpdatedAt,omitempty"`
}

//Domain defines the Domain address of a target.
type Domain struct {
	ID   string `json:"-" bson:"_id,omitempty"`
	Name string `json:"name,omitempty" bson:"Name"`
	IPs  []IP   `json:"ips,omitempty" bson:"IPs,omitempty"`

	CreatedAt int64 `json:"created_at" bson:"CreatedAt,omitempty"`
	UpdatedAt int64 `json:"updated_at" bson:"UpdatedAt,omitempty"`
}

//Project is the top level struct for a target. It contains a list of IPs and other metadata.
type Project struct {
	ID          string   `json:"-" bson:"_id,omitempty"`
	Name        string   `json:"name" bson:"Name"`
	Description string   `json:"description" bson:"Description,omitempty"`
	IPs         []IP     `json:"ips,omitempty" bson:"IPs,omitempty"`
	Domains     []Domain `json:"domains,omitempty" bson:"Domains,omitempty"`

	CreatedAt int64 `json:"created_at" bson:"CreatedAt,omitempty"`
	UpdatedAt int64 `json:"updated_at" bson:"UpdatedAt,omitempty"`
}

type User struct {
	ID       string `json:"-" bson:"_id,omitempty"`
	Name     string `json:"name" bson:"Name"`
	Password string `json:"password,omitempty"`
	Token    string `json:"token,omitempty" toml:"token"`

	CreatedAt int64 `json:"created_at" bson:"CreatedAt,omitempty"`
	UpdatedAt int64 `json:"updated_at" bson:"UpdatedAt,omitempty"`
}

//GenerateNewToken return a new random token string
func GenerateNewToken() string {
	b := make([]byte, 20)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

//Raws is a map, the key is module name (string). Each module write in a map of interface (key : string version of current timestamp)
type Raws map[string]map[string]interface{}

//AllRaws represents all Raws output, for a project string
type AllRaws map[string]Raws

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

	// IP
	CreateOrUpdateIP(projectName string, ip IP) error
	CreateOrUpdateIPs(projectName string, ip []IP) error
	GetIPs(projectName string) ([]IP, error)
	GetIP(projectName string, ip string) (IP, error)

	// Domain
	CreateOrUpdateDomain(projectName string, domain Domain) error
	CreateOrUpdateDomains(projectName string, domain []Domain) error
	GetDomains(projectName string) ([]Domain, error)
	GetDomain(projectName string, domain string) (Domain, error)

	// Port
	CreateOrUpdatePort(projectName string, ip string, port Port) error
	CreateOrUpdatePorts(projectName string, ip string, port []Port) error
	GetPorts(projectName string, ip string) ([]Port, error)
	GetPort(projectName string, ip string, port string) (Port, error)

	// URI (directory and files)
	CreateOrUpdateURI(projectName string, ip string, port string, dir URI) error
	CreateOrUpdateURIs(projectName string, ip string, port string, dir []URI) error
	GetURIs(projectName string, ip string, port string) ([]URI, error)
	GetURI(projectName string, ip string, port string, dir string) (URI, error)

	// Raw data
	AppendRawData(projectName string, moduleName string, data interface{}) error
	GetRaws(projectName string) (Raws, error)
	GetRawModule(projectName string, moduleName string) (map[string]interface{}, error)
}
