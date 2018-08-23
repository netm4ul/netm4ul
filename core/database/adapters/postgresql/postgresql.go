package postgresql

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/netm4ul/netm4ul/core/config"
	"github.com/netm4ul/netm4ul/core/database/models"

	pgdb "github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

/*
* This adapters relies on the "pg" package and ORM.
* "Models" are being extended in this package (file models.go)
 */

const DB_NAME = "netm4ul"

// PostgreSQL is the structure representing this adapter
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
		&pgRaw{},
	}

	for _, model := range reqs {
		log.Debugf("Creating table : %+v", model)
		err := pg.db.CreateTable(model, &orm.CreateTableOptions{
			FKConstraints: true,
			IfNotExists:   true,
		})
		if err != nil {
			return errors.New("Could not create table : " + err.Error())
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
	var p models.Project

	err := pg.db.Model(&p).Where("name = ?", project.Name).Select()
	// The project doesn't exist yet
	if err == pgdb.ErrNoRows {
		_, err := pg.db.Model(&p).Insert(&project)
		if err != nil {
			return errors.New("Could not insert project : " + err.Error())
		}
		return nil
	}

	// handle other errors
	if err != nil {
		return errors.New("Could not select project : " + err.Error())
	}
	// update if the project was found
	_, err = pg.db.Model(&project).Where("name = ?", project.Name).Returning("*").Update()
	if err != nil {
		return errors.New("Could not save project in the database :" + err.Error())
	}

	return nil
}

func (pg *PostgreSQL) getProjects() ([]pgProject, error) {
	var projects []pgProject

	err := pg.db.Model(&projects).Select()
	if err != nil {
		return nil, errors.New("Could not select projects : " + err.Error())
	}

	return projects, nil
}

func (pg *PostgreSQL) GetProjects() ([]models.Project, error) {
	var projects []models.Project

	ps, err := pg.getProjects()
	if err != nil {
		return nil, err
	}

	// convert to model
	for _, p := range ps {
		projects = append(projects, p.ToModel())
	}
	return projects, nil
}

func (pg *PostgreSQL) getProject(projectName string) (pgProject, error) {
	var project pgProject

	err := pg.db.Model(&project).Where("name = ?", projectName).First()
	if err != nil {
		return project, errors.New("Could not select project : " + err.Error())
	}

	return project, nil
}

func (pg *PostgreSQL) GetProject(projectName string) (models.Project, error) {
	p, err := pg.getProject(projectName)
	if err != nil {
		return models.Project{}, err
	}

	return p.ToModel(), nil
}

// IP
func (pg *PostgreSQL) createOrUpdateIP(projectName string, ip pgIP) error {
	log.Debugf("Inserting ip : %+v", ip)
	err := pg.db.Insert(&ip)

	if err != nil {
		return errors.New("Could not save ip in the database :" + err.Error())
	}
	return nil
}

func (pg *PostgreSQL) CreateOrUpdateIP(projectName string, ip models.IP) error {

	// convert to pgIP first
	pip := pgIP{}
	pip.FromModel(ip)

	err := pg.createOrUpdateIP(projectName, pip)
	if err != nil {
		return err
	}

	return nil
}

func (pg *PostgreSQL) createOrUpdateIPs(projectName string, ips []pgIP) error {
	for _, ip := range ips {
		err := pg.createOrUpdateIP(projectName, ip)
		if err != nil {
			return errors.New("Could not create or update ips : " + err.Error())
		}
	}
	return nil
}

func (pg *PostgreSQL) CreateOrUpdateIPs(projectName string, ips []models.IP) error {

	pips := []pgIP{}
	for _, ip := range ips {
		// convert ip
		pip := pgIP{}
		pip.FromModel(ip)
	}

	err := pg.createOrUpdateIPs(projectName, pips)
	if err != nil {
		return err
	}
	return nil
}

func (pg *PostgreSQL) getIPs(projectName string) ([]pgIP, error) {

	pgips := []pgIP{}
	err := pg.db.Model(&pgips).
		Where("projects.name = ?", projectName).
		Select()
	if err != nil {
		return nil, errors.New("Could not get IPs : " + err.Error())
	}

	return pgips, nil
}

func (pg *PostgreSQL) GetIPs(projectName string) ([]models.IP, error) {

	pgips, err := pg.getIPs(projectName)
	if err != nil {
		return nil, err
	}

	// convert back to the standard model
	ips := []models.IP{}
	for _, ip := range pgips {
		ips = append(ips, ip.ToModel())
	}

	return ips, nil
}

func (pg *PostgreSQL) getIP(projectName string, ip string) (pgIP, error) {

	pgip := pgIP{}
	err := pg.db.Model(&pgip).
		Where("projects.name = ?", projectName).
		Where("ips.value = ?", ip).
		Select()

	if err == sql.ErrNoRows {
		return pgIP{}, nil
	}
	if err != nil {
		return pgIP{}, errors.New("Could not get IP : " + err.Error())
	}

	return pgip, nil
}

func (pg *PostgreSQL) GetIP(projectName string, ip string) (models.IP, error) {

	pgip, err := pg.getIP(projectName, ip)
	if err != nil {
		return models.IP{}, err
	}
	return pgip.ToModel(), nil
}

// Domain
func (pg *PostgreSQL) createOrUpdateDomain(projectName string, domain pgDomain) error {

	err := pg.db.Insert(&domain)
	if err != nil {
		return errors.New("Could not save or update domain : " + err.Error())
	}

	return nil
}

func (pg *PostgreSQL) CreateOrUpdateDomain(projectName string, domain models.Domain) error {

	d := pgDomain{}
	d.FromModel(domain)

	err := pg.createOrUpdateDomain(projectName, d)
	if err != nil {
		return err
	}

	return nil
}

func (pg *PostgreSQL) createOrUpdateDomains(projectName string, domains []pgDomain) error {
	for _, domain := range domains {
		err := pg.createOrUpdateDomain(projectName, domain)
		if err != nil {
			return errors.New("Could not save or update domains : " + err.Error())
		}
	}
	return nil
}

func (pg *PostgreSQL) CreateOrUpdateDomains(projectName string, domains []models.Domain) error {

	pgds := []pgDomain{}
	for _, domain := range domains {
		pgd := pgDomain{}
		pgd.FromModel(domain)
		pgds = append(pgds, pgd)
	}

	err := pg.createOrUpdateDomains(projectName, pgds)
	if err != nil {
		return err
	}

	return nil
}

func (pg *PostgreSQL) getDomains(projectName string) ([]pgDomain, error) {
	domains := []pgDomain{}
	err := pg.db.Model(&domains).
		Where("projects.name = ?", projectName).
		Select()
	if err != nil {
		return nil, errors.New("Could not get domains : " + err.Error())
	}

	return domains, nil
}
func (pg *PostgreSQL) GetDomains(projectName string) ([]models.Domain, error) {

	domains := []models.Domain{}
	pgds, err := pg.getDomains(projectName)
	if err != nil {
		return nil, err
	}

	for _, d := range pgds {
		domains = append(domains, d.ToModel())
	}

	return domains, nil
}

func (pg *PostgreSQL) getDomain(projectName string, domainName string) (pgDomain, error) {
	domain := pgDomain{}

	err := pg.db.Model(&domain).
		Where("projects.name = ?", projectName).
		Where("domains.name = ?", domainName).
		Select()

	if err != nil {
		return pgDomain{}, errors.New("Could not get domain : " + err.Error())
	}

	return domain, nil
}

func (pg *PostgreSQL) GetDomain(projectName string, domainName string) (models.Domain, error) {
	d, err := pg.getDomain(projectName, domainName)
	if err != nil {
		return models.Domain{}, err
	}

	return d.ToModel(), nil
}

// Port
func (pg *PostgreSQL) createOrUpdatePort(projectName string, ip string, port pgPort) error {
	err := pg.db.Insert(&port)
	if err != nil {
		return errors.New("Could not create or update port : " + err.Error())
	}
	return nil
}

func (pg *PostgreSQL) CreateOrUpdatePort(projectName string, ip string, port models.Port) error {

	pgp := pgPort{}
	pgp.FromModel(port)
	err := pg.createOrUpdatePort(projectName, ip, pgp)
	if err != nil {
		return err
	}

	return nil
}

func (pg *PostgreSQL) createOrUpdatePorts(projectName string, ip string, ports []pgPort) error {
	for _, port := range ports {
		err := pg.createOrUpdatePort(projectName, ip, port)
		if err != nil {
			return err
		}
	}
	return nil
}

func (pg *PostgreSQL) CreateOrUpdatePorts(projectName string, ip string, ports []models.Port) error {
	for _, port := range ports {
		pgp := pgPort{}
		pgp.FromModel(port)
		err := pg.createOrUpdatePort(projectName, ip, pgp)
		if err != nil {
			return err
		}
	}
	return nil
}

func (pg *PostgreSQL) getPorts(projectName string, ip string) ([]pgPort, error) {

	ports := []pgPort{}

	err := pg.db.Model(&ports).Where("ips.value = ?", ip).Select()
	if err != nil {
		return nil, errors.New("Could not get ports : " + err.Error())
	}

	return ports, nil
}
func (pg *PostgreSQL) GetPorts(projectName string, ip string) ([]models.Port, error) {

	ports, err := pg.getPorts(projectName, ip)
	if err != nil {
		return nil, err
	}

	res := []models.Port{}
	for _, p := range ports {
		res = append(res, p.ToModel())
	}

	return res, nil
}

func (pg *PostgreSQL) getPort(projectName string, ip string, port string) (pgPort, error) {
	var p pgPort
	err := pg.db.Model(&p).
		Where("ips.value = ?", ip).
		Where("ports.number = ?", port).
		Select()

	if err != nil {
		return p, errors.New("Could not get port : " + err.Error())
	}

	return p, nil
}
func (pg *PostgreSQL) GetPort(projectName string, ip string, port string) (models.Port, error) {
	pgp, err := pg.getPort(projectName, ip, port)
	if err != nil {
		return models.Port{}, nil
	}
	return pgp.ToModel(), nil
}

// URI (directory and files)
func (pg *PostgreSQL) createOrUpdateURI(projectName string, ip string, port string, uri pgURI) error {
	//TOFIX
	err := pg.db.Insert(&uri)
	if err != nil {
		return err
	}
	return nil
}

func (pg *PostgreSQL) CreateOrUpdateURI(projectName string, ip string, port string, uri models.URI) error {

	puris := pgURI{}
	puris.FromModel(uri)

	err := pg.createOrUpdateURI(projectName, ip, port, puris)
	if err != nil {
		return err
	}

	return nil
}

func (pg *PostgreSQL) createOrUpdateURIs(projectName string, ip string, port string, uris []pgURI) error {
	// TOFIX
	// bulk insert!
	for _, uri := range uris {
		err := pg.createOrUpdateURI(projectName, ip, port, uri)
		if err != nil {
			return err
		}
	}
	return nil
}
func (pg *PostgreSQL) CreateOrUpdateURIs(projectName string, ip string, port string, uris []models.URI) error {
	puris := []pgURI{}
	for _, uri := range uris {
		puri := pgURI{}
		puri.FromModel(uri)
		puris = append(puris, puri)
	}

	err := pg.createOrUpdateURIs(projectName, ip, port, puris)
	if err != nil {
		return err
	}
	return nil
}

func (pg *PostgreSQL) getURIs(projectName string, ip string, port string) ([]pgURI, error) {

	uris := []pgURI{}

	err := pg.db.Model(&uris).
		Where("ips.value = ?", ip).
		Where("ports.number = ?", port).
		Select()
	if err != nil {
		return uris, errors.New("Could not get URIs : " + err.Error())
	}

	return nil, nil
}
func (pg *PostgreSQL) GetURIs(projectName string, ip string, port string) ([]models.URI, error) {

	uris := []models.URI{}

	puris, err := pg.getURIs(projectName, ip, port)
	if err != nil {
		return nil, err
	}

	for _, puri := range puris {
		uris = append(uris, puri.ToModel())
	}
	return uris, nil
}

func (pg *PostgreSQL) getURI(projectName string, ip string, port string, dir string) (pgURI, error) {

	uri := pgURI{}

	err := pg.db.Model(&uri).
		Where("ips.value = ?", ip).
		Where("ports.number = ?", port).
		Select()

	if err != nil {
		return pgURI{}, errors.New("Could not get URIs : " + err.Error())
	}

	return uri, nil
}

func (pg *PostgreSQL) GetURI(projectName string, ip string, port string, dir string) (models.URI, error) {

	uri, err := pg.getURI(projectName, ip, port, dir)
	if err != nil {
		return models.URI{}, err
	}

	return uri.ToModel(), err
}

// Raw data
func (pg *PostgreSQL) appendRawData(projectName string, raw pgRaw) error {
	err := pg.db.Insert(&raw)
	if err != nil {
		return errors.New("Could not insert raw : " + err.Error())
	}
	return nil
}

func (pg *PostgreSQL) AppendRawData(projectName string, raw models.Raw) error {
	praw := pgRaw{}
	praw.FromModel(raw)
	err := pg.appendRawData(projectName, praw)
	if err != nil {
		return err
	}
	return nil
}

func (pg *PostgreSQL) getRaws(projectName string) ([]pgRaw, error) {

	raws := []pgRaw{}

	err := pg.db.Model(&raws).Select()
	if err != nil {
		return nil, errors.New("Could not get raw : " + err.Error())
	}

	return raws, nil
}

func (pg *PostgreSQL) GetRaws(projectName string) ([]models.Raw, error) {

	raws := []models.Raw{}

	praws, err := pg.getRaws(projectName)
	if err != nil {
		return nil, err
	}

	for _, praw := range praws {
		raws = append(raws, praw.ToModel())
	}
	return raws, nil
}

func (pg *PostgreSQL) getRawModule(projectName string, moduleName string) (map[string][]pgRaw, error) {
	raws := []pgRaw{}
	err := pg.db.Model(raws).Where("raws.name = ?", projectName).Where("raws.moduleName = ?", moduleName).Select()

	if err != nil {
		return nil, errors.New("Could not get raw by module : " + err.Error())
	}
	var mapOfListOfRaw map[string][]pgRaw
	mapOfListOfRaw = make(map[string][]pgRaw)

	for _, r := range raws {
		mapOfListOfRaw[r.ModuleName] = append(mapOfListOfRaw[r.ModuleName], r)
	}
	return mapOfListOfRaw, nil
}

func (pg *PostgreSQL) GetRawModule(projectName string, moduleName string) (map[string][]models.Raw, error) {
	var mapOfListOfRaw map[string][]models.Raw
	mapOfListOfRaw = make(map[string][]models.Raw)

	rawsmap, err := pg.getRawModule(projectName, moduleName)

	if err != nil {
		return nil, err
	}

	//translate list of pgRaw to list of models.Raw
	for i, praws := range rawsmap {
		raws := []models.Raw{}
		for _, praw := range praws {
			raws = append(raws, praw.ToModel())
		}
		mapOfListOfRaw[i] = raws
	}

	return mapOfListOfRaw, nil
}
