package database

import (
	"log"

	mgo "gopkg.in/mgo.v2"
)

// GetProjects returns list of projects in the database. Take a session as argument
func GetProjects(session *mgo.Session) []Project {
	var p []Project
	err := session.DB(dbname).C("projects").Find(nil).All(&p)
	if err != nil {
		log.Println("Error in selecting projects", err)
		return nil
	}
	log.Println(p)
	return p
}
