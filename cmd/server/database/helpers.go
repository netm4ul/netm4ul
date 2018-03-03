package database

import (
	"fmt"
	"log"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// GetProjects returns list of projects in the database. Take a session as argument
func GetProjects(session *mgo.Session) []Project {
	var p []Project
	err := session.DB(dbname).C("projects").Find(nil).Select(bson.M{"Name": 1}).All(&p)
	if err != nil {
		log.Println("Error in selecting projects", err)
		return nil
	}
	fmt.Println("GetProjects p : ", p)
	return p
}

// GetProjectByName returns the project in the database by its name
func GetProjectByName(session *mgo.Session, name string) Project {
	var p Project
	err := session.DB(dbname).C("projects").Find(bson.M{"Name": name}).One(&p)
	if err != nil {
		log.Println("Error in selecting projects", err)
		return Project{}
	}
	fmt.Println("GetProjectByName p : ", p)
	return p
}
