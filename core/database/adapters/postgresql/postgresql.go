package postgresql

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/netm4ul/netm4ul/core/config"
	"github.com/netm4ul/netm4ul/core/database/models"

	pgdb "github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

/*
* This adapters relies on the "pg" ORM.
* Models are being extended in this package (file models.go)
*
 */
const DB_NAME = "netm4ul"

type PostgreSQL struct {
	cfg *config.ConfigToml
	db  *pgdb.DB
}

func InitDatabase(c *config.ConfigToml) *PostgreSQL {
	pg := PostgreSQL{}
	pg.cfg = c
	pg.Connect(c)
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
		&pgRaws{},
	}

	for _, model := range reqs {
		log.Debugf("Creating table : %+v", model)
		err := pg.db.CreateTable(model, &orm.CreateTableOptions{
			FKConstraints: true,
			IfNotExists:   true,
		})
		if err != nil {
			panic(err)
		}
	}
	return nil
}

func (pg *PostgreSQL) createDb() error {
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
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return err
	}
	// yep, configuration sqli, postgres limitation. cannot prepare this statement
	_, err = db.Exec(fmt.Sprintf(`create database %s`, strings.ToLower(pg.cfg.Database.Database)))
	db.Close()

	pg.db = pgdb.Connect(&pgdb.Options{
		Addr:     pg.cfg.Database.IP + ":" + strconv.Itoa(int(pg.cfg.Database.Port)),
		User:     pg.cfg.Database.User,
		Password: pg.cfg.Database.Password,
		Database: pg.cfg.Database.Database,
	})

	if err != nil {
		log.Error(err)
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

	pg.db = pgdb.Connect(&pgdb.Options{
		User:     c.Database.User,
		Password: c.Database.Password,
		Database: strings.ToLower(c.Database.Database),
		Addr:     c.Database.IP + ":" + strconv.Itoa(int(c.Database.Port)),
	})

	if pg.db == nil {
		return errors.New("Could not connect to the database : pg.db is nil")
	}

	return nil
}

//User

//TODO : refactor this.
func (pg *PostgreSQL) CreateOrUpdateUser(user models.User) error {

	tmpUser, err := pg.GetUser(user.Name)
	if err != nil {
		return errors.New("Could not get user : " + err.Error())
	}

	// transform user into pgUser model !
	pguser := pgUser{}
	pguser.FromModel(user)
	log.Debugf("pguser : %+v", &pguser)

	//user doesn't exist, create it
	if tmpUser.Name == "" {
		err = pg.db.Insert(&pguser)
		if err != nil {
			return errors.New("Could not insert user in the database : " + err.Error())
		}
		return nil
	}

	// update password
	if pguser.Password != "" && tmpUser.Password != pguser.Password {
		log.Debug("Updating password for user : ", pguser.Name)
		_, err = pg.db.Model(&pguser).
			Set("password = ?", pguser.Password).
			Where("name = ?", pguser.Name).
			Returning("*").
			Update()

		if err != nil {
			return errors.New("Couln't update user's password : " + err.Error())
		}
	}

	// update token
	if pguser.Token != "" && tmpUser.Token != pguser.Token {
		log.Debug("Updating token for user : ", pguser.Name)
		_, err = pg.db.Model(&pguser).
			Set("token = ?", pguser.Token).
			Where("name = ?", pguser.Name).
			Returning("*").
			Update()

		if err != nil {
			return errors.New("Couln't update user's token : " + err.Error())
		}
	}
	return nil
}

func (pg *PostgreSQL) GetUser(username string) (models.User, error) {

	pguser := pgUser{}
	err := pg.db.Model(&pguser).Where("name = ?", username).Select()

	user := pguser.ToModel()
	// Accept empty rows !
	if err != nil && err != pgdb.ErrNoRows {
		return user, errors.New("Could not get user by name : " + err.Error())
	}
	return user, nil
}

func (pg *PostgreSQL) GetUserByToken(token string) (models.User, error) {

	pguser := pgUser{}
	err := pg.db.Model(&pguser).Where("token = ?", token).Select()

	user := pguser.ToModel()

	if err != nil {
		return user, errors.New("Could not get user by token from the database : " + err.Error())
	}
	log.Errorf("User : %+v", user)

	return user, nil
}

/*
GenerateNewToken generates a new token and save it in the database.
It uses the function GenerateNewToken provided by the `models` class
*/
func (pg *PostgreSQL) GenerateNewToken(user models.User) error {

	user.Token = models.GenerateNewToken()
	err := pg.CreateOrUpdateUser(user)
	if err != nil {
		return errors.New("Could not generate a new token : " + err.Error())
	}
	return nil
}

//DeleteUser remove the user from the database (using its ID)
func (pg *PostgreSQL) DeleteUser(user models.User) error {
	err := pg.db.Delete(user)
	if err != nil {
		return errors.New("Could not delete user from the database : " + err.Error())
	}
	return nil
}

// Project
func (pg *PostgreSQL) CreateOrUpdateProject(project models.Project) error {

	_, err := pg.db.Model(&project).Returning("*").Update()
	if err != nil {
		return errors.New("Could not save project in the database :" + err.Error())
	}
	return nil
}

func (pg *PostgreSQL) GetProjects() ([]models.Project, error) {
	var projects []models.Project
	err := pg.db.Model(&projects).Select()

	if err != nil {
		return nil, errors.New("Could not select projects : " + err.Error())
	}

	return projects, nil
}

func (pg *PostgreSQL) GetProject(projectName string) (models.Project, error) {

	var project models.Project
	err := pg.db.Model(&project).Where("name = ?", projectName).First()
	if err != nil {
		return project, errors.New("Could not select project : " + err.Error())
	}

	return project, nil
}

// IP
func (pg *PostgreSQL) CreateOrUpdateIP(projectName string, ip models.IP) error {
	log.Debugf("Inserting ip : %+v", ip)
	err := pg.db.Insert(&ip)

	if err != nil {
		return errors.New("Could not save ip in the database :" + err.Error())
	}
	return nil
}

func (pg *PostgreSQL) CreateOrUpdateIPs(projectName string, ips []models.IP) error {
	for _, ip := range ips {
		err := pg.CreateOrUpdateIP(projectName, ip)
		if err != nil {
			return errors.New("Could not create or update ips : " + err.Error())
		}
	}
	return nil
}

func (pg *PostgreSQL) GetIPs(projectName string) ([]models.IP, error) {

	pgips := []pgIP{}
	err := pg.db.Model(&pgips).
		Where("projects.name = ?", projectName).
		Select()
	if err != nil {
		return nil, errors.New("Could not get IPs : " + err.Error())
	}

	// convert back to the standard model
	ips := []models.IP{}
	for _, ip := range pgips {
		ips = append(ips, ip.ToModel())
	}

	return ips, nil
}

func (pg *PostgreSQL) GetIP(projectName string, ip string) (models.IP, error) {
	pgip := pgIP{}
	err := pg.db.Model(&pgip).
		Where("projects.name = ?", projectName).
		Where("ips.value = ?", ip).
		Select()

	if err == sql.ErrNoRows {
		return models.IP{}, nil
	}
	if err != nil {
		return models.IP{}, errors.New("Could not get IP : " + err.Error())
	}

	return pgip.ToModel(), nil
}

// Domain
func (pg *PostgreSQL) CreateOrUpdateDomain(projectName string, domain models.Domain) error {
	err := pg.db.Insert(&domain)

	if err != nil {
		return errors.New("Could not save or update domain : " + err.Error())
	}
	return nil
}

func (pg *PostgreSQL) CreateOrUpdateDomains(projectName string, domains []models.Domain) error {

	//TODO bulk insert !
	for _, domain := range domains {
		err := pg.CreateOrUpdateDomain(projectName, domain)
		if err != nil {
			return errors.New("Could not save or update domains : " + err.Error())
		}
	}
	return nil
}

func (pg *PostgreSQL) GetDomains(projectName string) ([]models.Domain, error) {
	domains := []models.Domain{}
	err := pg.db.Model(&domains).
		Where("projects.name = ?", projectName).
		Select()
	if err != nil {
		return nil, errors.New("Could not get domains : " + err.Error())
	}

	return domains, nil
}

func (pg *PostgreSQL) GetDomain(projectName string, domainName string) (models.Domain, error) {
	domain := models.Domain{}

	err := pg.db.Model(&domain).
		Where("projects.name = ?", projectName).
		Where("domains.name = ?", domainName).
		Select()
	if err != nil {
		return models.Domain{}, errors.New("Could not get domain : " + err.Error())
	}

	return domain, nil
}

// Port
func (pg *PostgreSQL) CreateOrUpdatePort(projectName string, ip string, port models.Port) error {
	//TOFIX
	err := pg.db.Insert(&port)
	if err != nil {
		return errors.New("Could not create or update port : " + err.Error())
	}
	return nil
}

func (pg *PostgreSQL) CreateOrUpdatePorts(projectName string, ip string, ports []models.Port) error {
	for _, port := range ports {
		err := pg.CreateOrUpdatePort(projectName, ip, port)
		if err != nil {
			return err
		}
	}
	return nil
}

func (pg *PostgreSQL) GetPorts(projectName string, ip string) ([]models.Port, error) {
	ports := []models.Port{}
	err := pg.db.Model(&ports).Where("ips.value = ?", ip).Select()
	if err != nil {
		return nil, errors.New("Could not get ports : " + err.Error())
	}

	return ports, nil
}

func (pg *PostgreSQL) GetPort(projectName string, ip string, port string) (models.Port, error) {
	p := models.Port{}
	err := pg.db.Model(&p).
		Where("ips.value = ?", ip).
		Where("ports.number = ?", port).
		Select()

	if err != nil {
		return p, errors.New("Could not get port : " + err.Error())
	}
	return models.Port{}, nil
}

// URI (directory and files)
func (pg *PostgreSQL) CreateOrUpdateURI(projectName string, ip string, port string, dir models.URI) error {
	//TOFIX
	err := pg.db.Insert(&dir)
	if err != nil {
		return err
	}

	return nil
}

func (pg *PostgreSQL) CreateOrUpdateURIs(projectName string, ip string, port string, dirs []models.URI) error {
	// TOFIX
	// bulk insert!
	for _, dir := range dirs {
		err := pg.CreateOrUpdateURI(projectName, ip, port, dir)
		if err != nil {
			return err
		}
	}
	return nil
}

func (pg *PostgreSQL) GetURIs(projectName string, ip string, port string) ([]models.URI, error) {

	uris := []models.URI{}
	err := pg.db.Model(uris).
		Where("ips.value = ?", ip).
		Where("ports.number = ?", port).
		Select()

	if err != nil {
		return uris, errors.New("Could not get URIs : " + err.Error())
	}

	return uris, nil
}

func (pg *PostgreSQL) GetURI(projectName string, ip string, port string, dir string) (models.URI, error) {
	uri := models.URI{}
	err := pg.db.Model(uri).
		Where("ips.value = ?", ip).
		Where("ports.number = ?", port).
		Select()

	if err != nil {
		return uri, errors.New("Could not get URIs : " + err.Error())
	}

	return models.URI{}, nil
}

// Raw data
func (pg *PostgreSQL) AppendRawData(projectName string, moduleName string, data interface{}) error {

	jsonData, err := json.Marshal(data)
	if err != nil {
		return errors.New("Could not marshall data to json : " + err.Error())
	}
	r := models.Raws{Content: string(jsonData[:]), ModuleName: moduleName, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	err = pg.db.Insert(&r)

	if err != nil {
		return errors.New("Could not insert raw : " + err.Error())
	}

	return nil
}

func (pg *PostgreSQL) GetRaws(projectName string) (models.Raws, error) {
	raws := models.Raws{}
	err := pg.db.Model(raws).Select()
	if err != nil {
		return raws, errors.New("Could not get raw : " + err.Error())
	}
	return raws, nil
}

func (pg *PostgreSQL) GetRawModule(projectName string, moduleName string) (map[string]interface{}, error) {
	raws := map[string]interface{}{}
	err := pg.db.Model(raws).Select()

	if err != nil {
		return nil, errors.New("Could not get raw by module : " + err.Error())
	}

	return raws, nil
}
