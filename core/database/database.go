package database

import (
	"github.com/netm4ul/netm4ul/core/config"
	"github.com/netm4ul/netm4ul/core/database/models"
	"github.com/netm4ul/netm4ul/core/database/mongodb"
)

// NewDatabase returns the correct database adapter (mongodb, postegres...)
// TODO
func NewDatabase(c *config.ConfigToml) *models.Database {
	var db models.Database
	db = mongodb.InitDatabase(c)
	//todo load all database module and select the correct one from
	return &db
}
