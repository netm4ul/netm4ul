package models

import (
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
}

// Directory defines one directory from a remote target (webserver)
type Directory struct {
	ID   string `json:"-" bson:"_id,omitempty"`
	Name string `json:"name" bson:"Name"`
	Code string `json:"code,omitempty" bson:"Code,omitempty"`
}

// Port defines the basic structure for each port scanned on the target
type Port struct {
	ID          string      `json:"-" bson:"_id,omitempty"`
	Number      int16       `json:"number,omitempty" bson:"Number"`
	Protocol    string      `json:"protocol,omitempty" bson:"Protocol"`
	Status      string      `json:"status,omitempty" bson:"Status"` // open, filtered, closed
	Banner      string      `json:"banner,omitempty" bson:"Banner,omitempty"`
	Type        string      `json:"type,omitempty" bson:"Type,omitempty"`
	Directories []Directory `json:"value,omitempty" bson:"Value,omitempty"`
}

//IP defines the IP address of a target.
type IP struct {
	ID    string `json:"-" bson:"_id,omitempty"`
	Value string `json:"value,omitempty" bson:"Value"` // should be net.IP, but can't enforce that in the db...
	Ports []Port `json:"ports,omitempty" bson:"Ports,omitempty"`
}

//Project is the top level struct for a target. It contains a list of IPs and other metadata.
type Project struct {
	ID          string `json:"-" bson:"_id,omitempty"`
	Name        string `json:"name" bson:"Name"`
	Description string `json:"description" bson:"Description,omitempty"`
	UpdatedAt   int64  `json:"updated_at" bson:"UpdatedAt,omitempty"`
	IPs         []IP   `json:"ips,omitempty" bson:"IPs,omitempty"`
}

//Database is the mandatory interface for all custom database adapter
type Database interface {
	// General purpose functions
	Name() string
	SetupAuth(username, password, dbname string) error
	Connect(*config.ConfigToml)
	// Project
	CreateProject(projectName string)
	GetProjects() ([]Project, error)
	GetProjectByName(projectName string) (Project, error)
	UpdateProjectIPs(projectName string, ip IP)
	// IP
	AppendIP(ip IP)
	GetIPsByProjectName(projectName string) ([]IP, error)
	// Port
	AppendPorts(ports []Port)
	GetPortsByIP(projectName string, ip string) ([]Port, error)
	// Raw data
	AppendRawData(projectName string, moduleName string, data interface{})
}
