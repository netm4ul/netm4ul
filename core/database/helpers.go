package database

import (
	log "github.com/sirupsen/logrus"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// GetProjects returns list of projects in the database. Take a session as argument
func GetProjects(session *mgo.Session) []Project {
	var p []Project
	err := session.DB(DBname).C("projects").Find(nil).Select(bson.M{"Name": 1}).All(&p)
	if err != nil {
		log.Errorf("Error in selecting projects %+v", err)
		return nil
	}
	log.Infof("GetProjects p : %+v", p)
	return p
}

// GetProjectByName returns the project in the database by its name
func GetProjectByName(session *mgo.Session, name string) Project {
	var p Project
	err := session.DB(DBname).C("projects").Find(bson.M{"Name": name}).One(&p)
	if err != nil {
		log.Errorf("Error in selecting projects %+v", err)
		return Project{}
	}
	log.Infof("GetProjectByName p : %+v", p)
	return p
}
