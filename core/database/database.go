package database

import (
	"strings"

	"github.com/netm4ul/netm4ul/core/config"
	"github.com/netm4ul/netm4ul/core/database/adapters/jsondb"
	"github.com/netm4ul/netm4ul/core/database/adapters/mongodb"
	"github.com/netm4ul/netm4ul/core/database/models"
	log "github.com/sirupsen/logrus"
)

var adapters map[string]models.Database

func init() {
	adapters = make(map[string]models.Database, 3)
}

// NewDatabase returns the correct database adapter (mongodb, postegres...)
func NewDatabase(c *config.ConfigToml) models.Database {
	m := mongodb.InitDatabase(c)
	f := jsondb.InitDatabase(c)

	Register(m)
	Register(f)

	db := adapters[strings.ToLower(c.Database.DatabaseType)]
	log.Infof("Database list %+v, using %s from config file", adapters, c.Database.DatabaseType)

	return db
}

//Register : append the new database to the list of avaible connector
func Register(d models.Database) {
	adapters[strings.ToLower(d.Name())] = d
}
