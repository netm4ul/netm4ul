package postgresql

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/netm4ul/netm4ul/core/config"
	"github.com/netm4ul/netm4ul/core/database/models"

	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

const DB_NAME = "netm4ul"

type postgresql struct {
	cfg *config.ConfigToml
	db  *sql.DB
}

// General purpose functions
func (pg *postgresql) Name() string {
	return "PostgreSQL"
}

func (pg *postgresql) createTablesIfNotExist() error {

	if _, err := pg.db.Exec(createTableProjects); err != nil {
		return err
	}
	if _, err := pg.db.Exec(createTableIPs); err != nil {
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

func (pg *postgresql) SetupAuth(username, password, dbname string) error {
	return errors.New("Not implemented yet")
}

func (pg *postgresql) Connect(c *config.ConfigToml) error {
	db, err := sql.Open("postgres", fmt.Sprintf("user=%s dbname=%s host=%s sslmode=disable",
		c.Database.User, DB_NAME, c.Database.IP))
	if err != nil {
		return err
	}
	pg.db = db
	return nil
}

// Project
func (pg *postgresql) CreateOrUpdateProject(projectName string) error {

	var lastInsertID int
	err := pg.db.QueryRow(
		insertProject,
		projectName,
	).Scan(&lastInsertID)

	if err != nil {
		log.Error("Could not save project in the database :" + err.Error())
		return err
	}
	return nil
}

func (pg *postgresql) GetProjects() ([]models.Project, error) {
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

func (pg *postgresql) GetProject(projectName string) (models.Project, error) {

	row := pg.db.QueryRow(selectProjectByName, projectName)

	var id int64
	var name string
	var description string
	var updatedAt time.Time

	row.Scan(&id, &name, &description, &updatedAt)
	idstr := string(id)
	p := models.Project{ID: idstr, Name: name, Description: description, UpdatedAt: updatedAt.Unix()}

	return p, nil
}

// IP
func (pg *postgresql) CreateOrUpdateIP(projectName string, ip models.IP) error {
	var lastInsertID int
	err := pg.db.QueryRow(
		insertIP,
		ip.Value,
		projectName,
	).Scan(&lastInsertID)

	if err != nil {
		log.Error("Could not save project in the database :" + err.Error())
		return err
	}
	return nil
}

func (pg *postgresql) CreateOrUpdateIPs(projectName string, ips []models.IP) error {
	for _, ip := range ips {
		err := pg.CreateOrUpdateIP(projectName, ip)
		if err != nil {
			return err
		}
	}
	return nil
}

func (pg *postgresql) GetIPs(projectName string) ([]models.IP, error) {

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

func (pg *postgresql) GetIP(projectName string, ip string) (models.IP, error) {

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
func (pg *postgresql) CreateOrUpdatePort(projectName string, ip string, port models.Port) error {
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

func (pg *postgresql) CreateOrUpdatePorts(projectName string, ip string, ports []models.Port) error {
	for _, port := range ports {
		err := pg.CreateOrUpdatePort(projectName, ip, port)
		if err != nil {
			return err
		}
	}
	return nil
}

func (pg *postgresql) GetPorts(projectName string, ip string) ([]models.Port, error) {
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

func (pg *postgresql) GetPort(projectName string, ip string, port string) (models.Port, error) {
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
func (pg *postgresql) CreateOrUpdateURI(projectName string, ip string, port string, dir models.URI) error {
	var lastInsertID int64

	row := pg.db.QueryRow(insertURI, dir.Name, dir.Code, port, ip, projectName)
	err := row.Scan(lastInsertID)

	if err != nil {
		return err
	}

	return nil
}

func (pg *postgresql) CreateOrUpdateURIs(projectName string, ip string, port string, dirs []models.URI) error {
	for _, dir := range dirs {
		err := pg.CreateOrUpdateURI(projectName, ip, port, dir)
		if err != nil {
			return err
		}
	}
	return nil
}

func (pg *postgresql) GetURIs(projectName string, ip string, port string) ([]models.URI, error) {

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

func (pg *postgresql) GetURI(projectName string, ip string, port string, dir string) (models.URI, error) {
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
func (pg *postgresql) AppendRawData(projectName string, moduleName string, data interface{}) error {
	var lastInsertID int64
	row, err := pg.db.Query(insertRaw, moduleName, data, projectName)
	if err != nil {
		return err
	}

	err = row.Scan(&lastInsertID)
	if err != nil {
		return err
	}
	return nil
}

func (pg *postgresql) GetRaws(projectName string) (models.Raws, error) {
	var raws models.Raws
	return raws, errors.New("Not implemented yet")
}

func (pg *postgresql) GetRawModule(projectName string, moduleName string) (map[string]interface{}, error) {
	return nil, errors.New("Not implemented yet")
}
