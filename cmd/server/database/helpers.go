package database

import (
	"log"

	"github.com/netm4ul/netm4ul/cmd/cli"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// GetProjects returns list of projects in the database. Take a session as argument
func GetProjects(session *mgo.Session) []Project {
	var p []Project
	err := session.DB(DBname).C("projects").Find(nil).Select(bson.M{"Name": 1}).All(&p)
	if err != nil {
		log.Println(cli.Red("Error in selecting projects"), err)
		return nil
	}
	log.Printf(cli.Yellow("GetProjects p : %+v"), p)
	return p
}

// GetProjectByName returns the project in the database by its name
func GetProjectByName(session *mgo.Session, name string) Project {
	var p Project
	err := session.DB(DBname).C("projects").Find(bson.M{"Name": name}).One(&p)
	if err != nil {
		log.Printf(cli.Red("Error in selecting projects"), err)
		return Project{}
	}
	log.Printf(cli.Yellow("GetProjectByName p : %+v"), p)
	return p
}
