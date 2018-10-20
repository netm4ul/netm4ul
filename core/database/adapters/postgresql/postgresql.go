package postgresql

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/go-pg/pg/orm"
	"github.com/netm4ul/netm4ul/core/config"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

/*
 This adapters relies on the "pg" package and ORM.
 "Models" are being extended in this package (file models.go)
*/

const DB_NAME = "netm4ul"

// PostgreSQL is the structure representing this adapter
type PostgreSQL struct {
	cfg *config.ConfigToml
	db  *gorm.DB
}

func InitDatabase(c *config.ConfigToml) *PostgreSQL {
	pg := PostgreSQL{}
	pg.cfg = c
	pg.Connect(c)
	// pg.db.LogMode(true)
	return &pg
}

// General purpose functions

func (pg *PostgreSQL) Name() string {
	return "PostgreSQL"
}

func (pg *PostgreSQL) createTablesIfNotExist() error {
	log.Debug("Creating tables")

	reqs := []interface{}{
		&pgUser{},
		&pgProject{},
		&pgDomain{},
		&pgIP{},
		&pgPortType{},
		&pgPort{},
		&pgURI{},
		&pgRaw{},
		&userToProject{},
		&portToType{},
		&domainToIps{},
	}

	for _, model := range reqs {
		log.Debugf("Creating table : %+v", model)
		pg.db.CreateTable(model, &orm.CreateTableOptions{
			FKConstraints: true,
			IfNotExists:   true,
		})
	}
	return nil
}

func (pg *PostgreSQL) createDb() error {

	// uses default sql driver.
	// It's easier to create the database that way.
	log.Debug("Create database")
	strConn := fmt.Sprintf("user=%s host=%s password=%s sslmode=disable",
		pg.cfg.Database.User,
		pg.cfg.Database.IP,
		pg.cfg.Database.Password,
	)

	db, err := sql.Open("postgres", strConn)
	log.Debugf("StrConn : %s", strConn)
	if err != nil {
		return err
	}

	err = db.Ping()
	if err != nil {
		return err
	}

	// yep, configuration sqli, postgres limitation. cannot prepare this statement
	_, err = db.Exec(fmt.Sprintf(`create database %s`, strings.ToLower(pg.cfg.Database.Database)))
	db.Close()

	// close before reconnection ?
	pg.db.Close()
	err = pg.Connect(pg.cfg)
	if err != nil {
		return errors.New("Could not connect to the newly created database : " + err.Error())
	}

	log.Debugf("Database '%s' created !", strings.ToLower(pg.cfg.Database.Database))

	err = pg.createTablesIfNotExist()
	if err != nil {
		return errors.New("Could not create tables : " + err.Error())
	}
	return nil
}

//DeleteDatabase will drop all tables and remove the database from postgres
func (pg *PostgreSQL) DeleteDatabase() error {
	log.Debugf("DeleteDatabase postgres")

	strConn := fmt.Sprintf("user=%s host=%s password=%s sslmode=disable",
		pg.cfg.Database.User,
		pg.cfg.Database.IP,
		pg.cfg.Database.Password,
	)
	// Ensure all connections are closed before dropping the table
	pg.db.Close()

	db, err := sql.Open("postgres", strConn)
	log.Debugf("StrConn : %s", strConn)
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return err
	}

	// yep, configuration sqli, postgres limitation. cannot prepare this statement
	log.Infof("Dropping database : %s", fmt.Sprintf(dropDatabase, strings.ToLower(pg.cfg.Database.Database)))
	_, err = db.Exec(fmt.Sprintf(dropDatabase, strings.ToLower(pg.cfg.Database.Database)))
	if err != nil {
		return err
	}
	return nil
}

func (pg *PostgreSQL) SetupDatabase() error {
	log.Debugf("SetupDatabase postgres")
	err := pg.createDb()

	if err != nil {
		return errors.New("Could not setup the database : " + err.Error())
	}
	return nil
}

func (pg *PostgreSQL) SetupAuth(username, password, dbname string) error {
	log.Debugf("SetupAuth postgres")

	//TODO : create user/password
	// pg.Connect(pg.cfg)

	return nil
}

func (pg *PostgreSQL) Connect(c *config.ConfigToml) error {
	var err error
	log.Debugf("Connecting  to the database")
	strcon := fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=disable",
		c.Database.IP,
		c.Database.Port,
		c.Database.User,
		strings.ToLower(c.Database.Database),
		c.Database.Password,
	)
	pg.db, err = gorm.Open("postgres", strcon)

	if err != nil {
		return errors.New("Could not connect to the database : " + err.Error())
	}
	log.Debugf("Connected to the database")

	return nil
}
