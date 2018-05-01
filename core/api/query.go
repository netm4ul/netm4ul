package api

import (
	"github.com/netm4ul/netm4ul/core/database"
	"gopkg.in/mgo.v2/bson"
)

func (api *API) getIPsByProjectName(projectName string) ([]database.IP, error) {
	var project database.Project

	// maybe this should be inside a function ?
	sessionMgo := database.Connect()
	dbCollection := api.Session.Config.Database.Database
	c := sessionMgo.DB(dbCollection).C("projects")

	// query with "join" and ignoring ports (only getting project & ips)
	pipe := c.Pipe([]bson.M{
		{
			"$match": bson.M{
				"Name": projectName,
			},
		},
		{
			"$lookup": bson.M{
				"from":         "ips",
				"localField":   "IPs",
				"foreignField": "_id",
				"as":           "IPs",
			},
		},
		{
			"$limit": 1,
		},
		{
			"$project": bson.M{
				"IPs.Ports": 0,
			},
		},
	})

	err := pipe.One(&project)

	return project.IPs, err
}
