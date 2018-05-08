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
	log.Infof("Using database : %+v, and param %+v", adapters, c.Database.DatabaseType)
	log.Info("Using database : " + db.Name())

	return db
}

func Register(d models.Database) {
	adapters[strings.ToLower(d.Name())] = d
}
