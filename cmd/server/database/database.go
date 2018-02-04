package database

import (
	"fmt"
	"log"
	"time"

	mgo "gopkg.in/mgo.v2"
)

type Project struct {
	Name string
}

var db *mgo.Database

const dbname = "netm4ul"

// Connect to the database and return a session
func Connect() *mgo.Session {
	fmt.Println("Connecting !")
	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:    []string{"172.17.0.2"},
		Timeout:  60 * time.Second,
		Database: "NetM4ul",
		// TODO : security whatever ¯\_(ツ)_/¯
		// Username: "NetM4ul",
		// Password: "Password!",
	}
	session, err := mgo.DialWithInfo(mongoDBDialInfo)

	if err != nil {
		log.Fatal("Error connecting the database", err)
	}
	fmt.Println("Connected ! hurra")
	return session
}

// CreateProject create a new project structure inside db
func CreateProject(session *mgo.Session, projectName string) {
	// mongodb will create collection on use.
	fmt.Println("Should add " + projectName + "to the collections 'projects'")
	c := session.DB(dbname).C("projects")
	p := Project{Name: projectName}

	err := c.Insert(p)
	if err != nil {
		log.Fatal(err)
	}
}
