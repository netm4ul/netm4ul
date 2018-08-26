package database

import (
	"errors"
	"strings"

	"github.com/netm4ul/netm4ul/core/config"
	"github.com/netm4ul/netm4ul/core/database/adapters/jsondb"
	"github.com/netm4ul/netm4ul/core/database/adapters/postgresql"
	"github.com/netm4ul/netm4ul/core/database/adapters/testadapter"
	"github.com/netm4ul/netm4ul/core/database/models"
	log "github.com/sirupsen/logrus"
)

var adapters map[string]models.Database

func init() {
	adapters = make(map[string]models.Database, 0)
}

// NewDatabase returns the correct database adapter (mongodb, postegres...)
func NewDatabase(c *config.ConfigToml) (models.Database, error) {
	// m := mongodb.InitDatabase(c)
	f := jsondb.InitDatabase(c)
	p := postgresql.InitDatabase(c)
	t := testadapter.InitDatabase(c)

	// Register(m)
	Register(f)
	Register(p)
	Register(t)

	db, ok := adapters[strings.ToLower(c.Database.DatabaseType)]
	if !ok {
		keys := make([]string, 0, len(adapters))
		for k := range adapters {
			keys = append(keys, k)
		}

		return nil, errors.New("Could not use the database named '" + strings.ToLower(c.Database.DatabaseType) + "'. Have : " + "[" + strings.Join(keys, ", ") + "]")
	}

	log.Debugf("Database list %+v, using %s from config file", adapters, c.Database.DatabaseType)
	log.Debugf("New Database : %+v", db)
	return db, nil
}

//Register : append the new database to the list of avaible connector
func Register(d models.Database) {
	adapters[strings.ToLower(d.Name())] = d
}
