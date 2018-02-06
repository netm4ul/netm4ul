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

type Hop struct {
	IP  net.IP
	Max float32
	Min float32
	Avg float32
}

type Route struct {
	Source      string
	Destination string
	Hops        []Hop
}
type Directory struct {
	Name string
	Code string
}

type Port struct {
	Number      int16
	Protocol    string
	Status      string // open, filtered, closed
	Banner      string
	Type        string
	Directories []Directory
}
type IP struct {
	Value net.IP
	Ports []Port
}

type Project struct {
	Name    string
	Updated time.Time
	IPs     []interface{}
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
	fmt.Println("Should add " + projectName + "to the collections 'projects'")
	c := session.DB(dbname).C("projects")

	info, err := c.Upsert(bson.M{"Name": projectName}, bson.M{"$set": bson.M{"updatedAt": time.Now()}})
	log.Println(info)
	if err != nil {
		log.Fatal(err)
	}
}
