package database

import (
	"github.com/netm4ul/netm4ul/core/config"
	"github.com/netm4ul/netm4ul/core/database/files"
	"github.com/netm4ul/netm4ul/core/database/models"
	"github.com/netm4ul/netm4ul/core/database/mongodb"
)

var adapters map[string]models.Database

func init() {
	adapters = make(map[string]models.Database, 3)
}

// NewDatabase returns the correct database adapter (mongodb, postegres...)
// TODO
func NewDatabase(c *config.ConfigToml) models.Database {
	m := mongodb.InitDatabase(c)
	f := files.InitDatabase(c)

	Register(m)
	Register(f)

	db := adapters[c.Database.DatabaseType]

	return db
}

func Register(d models.Database) {
	adapters[d.Name()] = d
}
