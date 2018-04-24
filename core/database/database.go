package database

import (
	"fmt"
	"net"
	"time"

	"github.com/netm4ul/netm4ul/core/config"
	log "github.com/sirupsen/logrus"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

/*
	See README for the database "schema"
	This file contains some generic informations.
*/

// Hop defines each "hop" from the host (netm4ul client) to the target.
type Hop struct {
	ID  bson.ObjectId `bson:"_id,omitempty"`
	IP  net.IP
	Max float32
	Min float32
	Avg float32
}

// Route defines the route from the host (netm4ul client) to the target
type Route struct {
	ID          bson.ObjectId `bson:"_id,omitempty"`
	Source      string        `json:"source,omitempty" bson:"Source"`
	Destination string        `json:"destination,omitempty" bson:"Destination"`
	Hops        []Hop         `json:"hops,omitempty" bson:"Hops,omitempty"`
}

// Directory defines one directory from a remote target (webserver)
type Directory struct {
	ID   bson.ObjectId `bson:"_id,omitempty"`
	Name string        `json:"name" bson:"Name"`
	Code string        `json:"code,omitempty" bson:"Code,omitempty"`
}

// Port defines the basic structure for each port scanned on the target
type Port struct {
	ID          bson.ObjectId `bson:"_id,omitempty"`
	Number      int16         `json:"number,omitempty" bson:"Number"`
	Protocol    string        `json:"protocol,omitempty" bson:"Protocol"`
	Status      string        `json:"status,omitempty" bson:"Status"` // open, filtered, closed
	Banner      string        `json:"banner,omitempty" bson:"Banner,omitempty"`
	Type        string        `json:"type,omitempty" bson:"Type,omitempty"`
	Directories []Directory   `json:"value,omitempty" bson:"Value,omitempty"`
}

//IP defines the IP address of a target.
type IP struct {
	ID    bson.ObjectId `bson:"_id,omitempty"`
	Value net.IP        `json:"value,omitempty" bson:"Value"`
	Ports []Port        `json:"ports,omitempty" bson:"Ports,omitempty"`
}

//Project is the top level struct for a target. It contains a list of IPs and other metadata.
type Project struct {
	ID          bson.ObjectId `bson:"_id,omitempty"`
	Name        string        `json:"name" bson:"Name"`
	Description string        `json:"description" bson:"Description,omitempty"`
	UpdatedAt   int64         `json:"updated_at" bson:"UpdatedAt,omitempty"`
	IPs         []IP          `json:"ips" bson:"omitempty"`
}

var cfg *config.ConfigToml

func InitDatabase(c *config.ConfigToml) {
	cfg = c
}

// Connect to the database and return a session
func ConnectWithoutCreds() *mgo.Session {
	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:    []string{config.Config.Database.IP}, // array of ip (sharding & whatever), just 1 for now
		Timeout:  10 * time.Second,
		Database: "NetM4ul",
	}
	session, err := mgo.DialWithInfo(mongoDBDialInfo)

	if err != nil {
		log.Fatal("Error connecting with the database : %s", err.Error())
	}
	log.Info("Connected to the database : %s", config.Config.Database.IP)
	return session
}

// Connect to the database and return a session
func Connect() *mgo.Session {
	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:    []string{cfg.Database.IP}, // array of ip (sharding & whatever), just 1 for now
		Timeout:  10 * time.Second,
		Database: "NetM4ul",
		Username: cfg.Database.User,
		Password: cfg.Database.Password,
	}
	session, err := mgo.DialWithInfo(mongoDBDialInfo)

	if err != nil {
		log.Fatalf("Error connecting with the database : %s", err.Error())
	}
	log.Infof("Connected to the database : %s", cfg.Database.IP)
	return session
}

// CreateProject create a new project structure inside db
func CreateProject(session *mgo.Session, projectName string) {
	// mongodb will create collection on use.
	c := session.DB(cfg.Database.Collection).C("projects")

	info, err := c.Upsert(bson.M{"Name": projectName}, bson.M{"$set": bson.M{"UpdatedAt": time.Now().Unix()}})

	if cfg.Verbose && info.Updated == 1 {
		log.Infof("Info : %+v", info)
		log.Infof("Adding %s to the collections 'projects'", projectName)
	}

	if err != nil {
		log.Fatal(err)
	}
}

//UpsertRawData is used by module to store raw results into the database.
func UpsertRawData(session *mgo.Session, projectName string, data bson.M) {
	c := session.DB(cfg.Database.Collection).C("projects")
	info, err := c.Upsert(bson.M{"Name": projectName}, bson.M{"$push": data})
	log.Infof("Info : %+v", info)
	if err != nil {
		log.Fatal(err)
	}
}

// AppendIP is used by module to store result into the database
func appendIP(session *mgo.Session, ip IP) {
	fmt.Println("Adding IP to project facebook.com")
	c := session.DB(cfg.Database.Collection).C("projects")

	ip.ID = bson.NewObjectId()

	portsArr := make([]bson.ObjectId, len(ip.Ports))
	for k := range ip.Ports {
		fmt.Println("k:", k)
		fmt.Println("portsFunc[k]:", ip.Ports[k].ID)
		portsArr[k] = ip.Ports[k].ID
	}
	fmt.Println("portsArr : ", portsArr)

	_, err := c.Upsert(bson.M{"Name": "projects"}, bson.M{"_id": ip.ID, "Value": ip.Value, "Ports": portsArr})
	if err != nil {
		log.Fatal("Something went wrong : ", err)
	}
}

func appendPorts(session *mgo.Session, ports []Port) {
	c := session.DB(cfg.Database.Collection).C("projects")

	for v := range ports {
		info, err := c.Upsert(bson.M{"Name": "projects"},
			bson.M{"_id": ports[v].ID, "Number": ports[v].Number,
				"Protocol": ports[v].Protocol,
				"Status":   ports[v].Status,
				"Banner":   ports[v].Banner})
		if err != nil {
			log.Fatal("Something went wrong (PORT) : ", err)
		}
		fmt.Println(info)
		v++
	}
}
