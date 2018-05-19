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

	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

const DB_NAME = "netm4ul"

type PostgreSQL struct {
	cfg *config.ConfigToml
	db  *sql.DB
}

func InitDatabase(c *config.ConfigToml) *PostgreSQL {
	pg := PostgreSQL{}
	pg.cfg = c
	return &pg
}

// General purpose functions

func (pg *PostgreSQL) Name() string {
	return "PostgreSQL"
}

func (pg *PostgreSQL) createTablesIfNotExist() error {

	if _, err := pg.db.Exec(createTableProjects); err != nil {
		return err
	}
	if _, err := pg.db.Exec(createTableIPs); err != nil {
		return err
	}
	if _, err := pg.db.Exec(createTablePortTypes); err != nil {
		return err
	}
	if _, err := pg.db.Exec(createTablePorts); err != nil {
		return err
	}
	if _, err := pg.db.Exec(createTableURIs); err != nil {
		return err
	}
	if _, err := pg.db.Exec(createTableRaws); err != nil {
		return err
	}
	return nil
}

func (pg *PostgreSQL) createDb() {
	log.Debug("Create database")
	strConn := fmt.Sprintf("user=%s host=%s password=%s sslmode=disable",
		pg.cfg.Database.User,
		pg.cfg.Database.IP,
		pg.cfg.Database.Password,
	)
	db, err := sql.Open("postgres", strConn)
	log.Debugf("StrConn : %s", strConn)
	if err != nil {
		log.Error(err)
	}
	// yep, configuration sqli, postgres limitation. cannot prepare this statement
	_, err = db.Exec(fmt.Sprintf(`create database %s`, strings.ToLower(pg.cfg.Database.Database)))

	// we ignore if the database already exist
	if err != nil {
		log.Error(err)
	}
	log.Debug("Database created !")

}

func (pg *PostgreSQL) SetupAuth(username, password, dbname string) error {
	log.Debugf("SetupAuth postgres")
	pg.createDb()
	pg.Connect(pg.cfg)
	//TODO : create user/password
	err := pg.createTablesIfNotExist()
	if err != nil {
		return err
	}

	return nil
}

func (pg *PostgreSQL) Connect(c *config.ConfigToml) error {

	connStr := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%d sslmode=disable", c.Database.User, c.Database.Password, strings.ToLower(c.Database.Database), c.Database.IP, c.Database.Port)
	log.Debugf("Connection string : %s", connStr)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	pg.db = db
	return nil
}

//User
func (pg *PostgreSQL) CreateOrUpdateUser(user models.User) error {
	return errors.New("Not implemented yet")
}

func (pg *PostgreSQL) GetUser(username string) (models.User, error) {
	return models.User{}, errors.New("Not implemented yet")
}

func (pg *PostgreSQL) GenerateNewToken(user models.User) error {
	return errors.New("Not implemented yet")
}

func (pg *PostgreSQL) DeleteUser(user models.User) error {
	return errors.New("Not implemented yet")
}

// Project
func (pg *PostgreSQL) CreateOrUpdateProject(project models.Project) error {

	var lastInsertID int
	err := pg.db.QueryRow(
		insertProject,
		project.Name,
		project.Description,
	).Scan(&lastInsertID)

	if err != nil {
		log.Error("Could not save project in the database :" + err.Error())
		return err
	}
	return nil
}

func (pg *PostgreSQL) GetProjects() ([]models.Project, error) {
	rows, err := pg.db.Query(selectProjects)
	if err != nil {
		return nil, err
	}

	var id int64
	var name string
	var description string
	var updatedAt time.Time
	projects := []models.Project{}

	for rows.Next() {
		rows.Scan(&id, &name, &description, &updatedAt)
		idstr := string(id)
		p := models.Project{ID: idstr, Name: name, Description: description, UpdatedAt: updatedAt.Unix()}
		projects = append(projects, p)
	}

	return projects, nil
}

func (pg *PostgreSQL) GetProject(projectName string) (models.Project, error) {

	row := pg.db.QueryRow(selectProjectByName, projectName)

	var id int64
	var name string
	var description string
	var updatedAt time.Time

	row.Scan(&id, &name, &description, &updatedAt)
	idstr := strconv.Itoa(int(id))
	p := models.Project{ID: idstr, Name: name, Description: description, UpdatedAt: updatedAt.Unix()}

	return p, nil
}

// IP
func (pg *PostgreSQL) CreateOrUpdateIP(projectName string, ip models.IP) error {
	var lastInsertID int
	err := pg.db.QueryRow(
		insertIP,
		ip.Value,
		projectName,
	).Scan(&lastInsertID)

	if err != nil {
		log.Error("Could not save ip in the database :" + err.Error())
		return err
	}
	return nil
}

func (pg *PostgreSQL) CreateOrUpdateIPs(projectName string, ips []models.IP) error {
	for _, ip := range ips {
		err := pg.CreateOrUpdateIP(projectName, ip)
		if err != nil {
			return err
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
