package database

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/netm4ul/netm4ul/cmd/config"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
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
	Source      string
	Destination string
	Hops        []Hop
}

// Directory defines one directory from a remote target (webserver)
type Directory struct {
	Name string
	Code string
}

// Port defines the basic structure for each port scanned on the target
type Port struct {
	Number      int16
	Protocol    string
	Status      string // open, filtered, closed
	Banner      string
	Type        string
	Directories []Directory
}

//IP defines the IP address of a target.
type IP struct {
	Value net.IP
	Ports []Port
}

//Project is the top level struct for a target. It contains a list of IPs and other metadata.
type Project struct {
	Name      string
	UpdatedAt int64
	IPs       []interface{}
}

var db *mgo.Database

const dbname = "netm4ul"

// Connect to the database and return a session
func Connect() *mgo.Session {
	fmt.Println("Connecting !")
	log.Println("config.Config.Database.IP : ", config.Config.Database.IP)
	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:    []string{config.Config.Database.IP}, // array of ip (sharding & whatever), just 1 for now
		Timeout:  10 * time.Second,
		Database: "NetM4ul",
		// TODO : security whatever ¯\_(ツ)_/¯
		// Username: "NetM4ul",
		// Password: "Password!",
	}
	session, err := mgo.DialWithInfo(mongoDBDialInfo)

	if err != nil {
		log.Fatal("Error connecting with the database", err)
	}
	fmt.Println("Connected !")
	return session
}

// CreateProject create a new project structure inside db
func CreateProject(session *mgo.Session, projectName string) {
	// mongodb will create collection on use.
	fmt.Println("Should add " + projectName + " to the collections 'projects'")
	c := session.DB(dbname).C("projects")

	info, err := c.Upsert(bson.M{"Name": projectName}, bson.M{"$set": bson.M{"UpdatedAt": time.Now().Unix()}})
	log.Println(info)
	if err != nil {
		log.Fatal(err)
	}
}
