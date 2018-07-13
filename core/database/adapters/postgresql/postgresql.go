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

	reqs := []interface{}{
		&pgUser{},
		&pgProject{},
		&pgDomain{},
		&pgIP{},
		&pgPortType{},
		&pgPort{},
		&pgURI{},
		// &pgRaw{},
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
		Addr:     pg.cfg.Database.IP,
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
	User := new(models.User)

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
		Addr:     c.Database.IP + strconv.Itoa(int(c.Database.Port)),
	})

	if pg.db == nil {
		return errors.New("Could not connect to the database : pg.db is nil")
	}

	return nil
}

//User

//TODO : refactor this.
func (pg *PostgreSQL) CreateOrUpdateUser(user pgUser) error {

	u, err := pg.GetUser(user.Name)
	if err != nil {
		return errors.New("Error : " + err.Error())
	}

	//user exist, update it
	if u.ID != 0 {
		// update password
		if user.Password != "" && u.Password != user.Password {
			log.Debug("Updating password for user", user.Name)
			_, err = pg.db.Model(user).
				Set("password = ?password").
				WherePK().
				Returning("*").
				Update()
			if err != nil {
				return errors.New("pq error: " + err.Error())
			}
		}
		// update token
		if user.Token != "" && u.Token != user.Token {
			log.Debug("Updating token for user", user.Name)
			_, err = pg.db.Model(user).
				Set("password = ?password").
				WherePK().
				Returning("*").
				Update()

			if err != nil {
				return errors.New("pq error: " + err.Error())
			}
		}
		return nil
	}

	err = pg.db.Insert(user)
	if err != nil {
		return errors.New("Could not insert user in the database : " + err.Error())
	}

	return nil
}

func (pg *PostgreSQL) GetUser(username string) (pgUser, error) {

	user := pgUser{}
	err := pg.db.Model(user).Where("username = ?username", username).Select()

	// Accept empty rows !
	if err != nil && err != sql.ErrNoRows {
		return user, errors.New("Could not get user by name : " + err.Error())
	}
	return user, nil
}

func (pg *PostgreSQL) GetUserByToken(token string) (pgUser, error) {

	user := pgUser{}
	err := pg.db.Model(user).Where("token = ?token", token).Select()

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

	_, err := pg.db.Model(project).Returning("*").Update()
	if err != nil {
		return errors.New("Could not save project in the database :" + err.Error())
	}
	return nil
}

func (pg *PostgreSQL) GetProjects() ([]models.Project, error) {
	var projects []models.Project
	err := pg.db.Model(projects).Select()

	if err != nil {
		return nil, errors.New("Could not select projects : " + err.Error())
	}

	return projects, nil
}

func (pg *PostgreSQL) GetProject(projectName string) (models.Project, error) {

	var project models.Project
	err := pg.db.Model(project).Where("name = ?", projectName).First()
	if err != nil {
		return project, errors.New("Could not select project : " + err.Error())
	}

	return project, nil
}

// IP
func (pg *PostgreSQL) CreateOrUpdateIP(projectName string, ip models.IP) error {

	err := pg.db.Model(ip)
	// 	ip.Value,
	// 	projectName,
	// ).Scan(&lastInsertID)

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

	ips := []models.IP{}
	rows, err := pg.db.Query(selectIPsByProjectName, projectName)
	if err != nil {
		return nil, err
	}

	var id int64
	var value string

	for rows.Next() {
		rows.Scan(&id, &value)
		idstr := string(id)
		ip := models.IP{ID: idstr, Value: value}
		ips = append(ips, ip)
	}

	return ips, nil
}

func (pg *PostgreSQL) GetIP(projectName string, ip string) (models.IP, error) {

	row := pg.db.QueryRow(selectIPByProjectName, projectName, ip)

	var id int64
	var value string

	err := row.Scan(&id, &value)

	if err == sql.ErrNoRows {
		return models.IP{}, nil
	}
	if err != nil {
		return models.IP{}, err
	}

	idstr := string(id)
	result := models.IP{ID: idstr, Value: value}

	return result, nil
}

// Domain
func (pg *PostgreSQL) CreateOrUpdateDomain(projectName string, domain models.Domain) error {
	var lastInsertID int
	err := pg.db.QueryRow(insertDomain, domain.Name, projectName).Scan(lastInsertID)

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
	rows, err := pg.db.Query(selectDomains, projectName)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		domain := models.Domain{}
		err = rows.Scan(&domain.Name, &domain.CreatedAt, &domain.UpdatedAt)
		if err != nil {
			return nil, err
		}

		domains = append(domains, domain)
	}

	return domains, nil
}

func (pg *PostgreSQL) GetDomain(projectName string, domainName string) (models.Domain, error) {
	domain := models.Domain{}
	err := pg.db.QueryRow(selectDomain, domainName, projectName).Scan(
		&domain.ID,
		&domain.Name,
		&domain.CreatedAt,
		&domain.UpdatedAt,
	)

	if err != nil {
		return models.Domain{}, errors.New("Could not get domain : " + err.Error())
	}

	return domain, nil
}

// Port
func (pg *PostgreSQL) CreateOrUpdatePort(projectName string, ip string, port models.Port) error {
	var lastInsertID int
	err := pg.db.QueryRow(insertPort,
		port.Number,
		port.Protocol,
		port.Status,
		port.Banner,
		port.Type,
		projectName,
		ip,
	).Scan(lastInsertID)

	if err != nil {
		return err
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
	rows, err := pg.db.Query(selectPortsByProjectNameAndIP, projectName, ip)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		p := models.Port{}
		err = rows.Scan(&p.ID, &p.Number, &p.Protocol, &p.Status, &p.Banner, &p.Type)
		if err != nil {
			return nil, err
		}

		ports = append(ports, p)
	}

	return ports, nil
}

func (pg *PostgreSQL) GetPort(projectName string, ip string, port string) (models.Port, error) {
	ports, err := pg.GetPorts(projectName, ip)
	if err != nil {
		return models.Port{}, nil
	}

	for _, p := range ports {
		if strconv.Itoa(int(p.Number)) == port {
			return p, nil
		}
	}
	//not found
	return models.Port{}, nil
}

// URI (directory and files)
func (pg *PostgreSQL) CreateOrUpdateURI(projectName string, ip string, port string, dir models.URI) error {
	var lastInsertID int64

	row := pg.db.QueryRow(insertURI, dir.Name, dir.Code, port, ip, projectName)
	err := row.Scan(lastInsertID)

	if err != nil {
		return err
	}

	return nil
}

func (pg *PostgreSQL) CreateOrUpdateURIs(projectName string, ip string, port string, dirs []models.URI) error {
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
	rows, err := pg.db.Query(selectURIsByProjectNameAndIPAndPort, projectName, port)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		uri := models.URI{}
		err := rows.Scan(&uri.ID, &uri.Name, &uri.Code)
		if err != nil {
			return nil, err
		}
		uris = append(uris, uri)
	}

	return uris, nil
}

func (pg *PostgreSQL) GetURI(projectName string, ip string, port string, dir string) (models.URI, error) {
	uris, err := pg.GetURIs(projectName, ip, port)
	if err != nil {
		return models.URI{}, err
	}
	for _, uri := range uris {
		if uri.Name == dir {
			return uri, nil
		}
	}
	//not found
	return models.URI{}, nil
}

// Raw data
func (pg *PostgreSQL) AppendRawData(projectName string, moduleName string, data interface{}) error {
	var lastInsertID int64

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	row := pg.db.QueryRow(insertRaw, moduleName, jsonData, projectName)
	err = row.Scan(&lastInsertID)
	if err != nil {
		return err
	}

	return nil
}

func (pg *PostgreSQL) GetRaws(projectName string) (models.Raws, error) {
	raws := models.Raws{}
	rows, err := pg.db.Query(selectRawsByProjectName, projectName)
	if err != nil {
		return models.Raws{}, err
	}
	var id int64
	var module string
	var project string
	var data interface{}
	var createdAt time.Time

	for rows.Next() {
		err = rows.Scan(&id, &module, &project, &data, &createdAt)
		if err != nil {
			return models.Raws{}, err
		}

		raws[module] = make(map[string]interface{})
		raws[module][strconv.Itoa(int(createdAt.Unix()))] = data
	}
	return raws, nil
}

func (pg *PostgreSQL) GetRawModule(projectName string, moduleName string) (map[string]interface{}, error) {
	raws := map[string]interface{}{}
	rows, err := pg.db.Query(selectRawsByProjectNameAndModuleName, projectName, moduleName)

	if err != nil {
		return nil, err
	}

	var id int64
	var module string
	var project string
	var data interface{}
	var createdAt time.Time

	for rows.Next() {
		err = rows.Scan(&id, &module, &project, &data, &createdAt)
		if err != nil {
			return nil, err
		}

		raws[strconv.Itoa(int(createdAt.Unix()))] = data
	}
	return raws, nil
}
