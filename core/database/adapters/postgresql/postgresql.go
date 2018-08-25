package postgresql

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/go-pg/pg/orm"
	"github.com/netm4ul/netm4ul/core/config"
	"github.com/netm4ul/netm4ul/core/database/models"

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

//User

//TODO : refactor this.
func (pg *PostgreSQL) CreateOrUpdateUser(user models.User) error {

	tmpUser, err := pg.getUser(user.Name)
	if err != nil {
		return errors.New("Could not get user : " + err.Error())
	}

	// transform user into pgUser model !
	pguser := pgUser{}
	pguser.FromModel(user)
	log.Debugf("pguser : %+v", &pguser)
	log.Debugf("user : %+v", &user)

	//user doesn't exist, create it
	if tmpUser.Name == "" {
		res := pg.db.Create(&pguser)
		if res.Error != nil {
			return errors.New("Could not insert user in the database : " + res.Error.Error())
		}
		return nil
	}

	// update password
	if pguser.Password != "" && tmpUser.Password != pguser.Password {
		log.Debug("Updating password for user : ", pguser.Name)
		tmpUser.Password = pguser.Password
	}

	//if the in-database user doesn't have a token, create one : it might be a security issues without one.
	if tmpUser.Token == "" {
		tmpUser.Token = models.GenerateNewToken()
	}

	if pguser.Token != "" && tmpUser.Token != pguser.Token {
		log.Debug("Updating token for user : ", pguser.Name)
		tmpUser.Token = pguser.Token
	}

	log.Debugf("Writing tmp user : %+v", tmpUser)
	res := pg.db.Model(&tmpUser).Update(&tmpUser)

	if res.Error != nil {
		return errors.New("Could not update user : " + res.Error.Error())
	}
	return nil
}

func (pg *PostgreSQL) getUser(username string) (pgUser, error) {

	pguser := pgUser{}
	res := pg.db.Where("name = ?", username).First(&pguser)

	// Accept empty rows !
	if res.Error != nil && !gorm.IsRecordNotFoundError(res.Error) {
		return pgUser{}, errors.New("Could not get user by name : " + res.Error.Error())
	}
	return pguser, nil
}

func (pg *PostgreSQL) GetUser(username string) (models.User, error) {
	pguser, err := pg.getUser(username)
	if err != nil {
		return models.User{}, err
	}
	return pguser.ToModel(), err
}

func (pg *PostgreSQL) getUserByToken(token string) (pgUser, error) {

	pguser := pgUser{}
	res := pg.db.Where("token = ?", token).First(&pguser)
	// Accept empty rows !
	if res.Error != nil && !gorm.IsRecordNotFoundError(res.Error) {
		return pgUser{}, errors.New("Could not get user by token : " + res.Error.Error())
	}
	return pguser, nil
}

func (pg *PostgreSQL) GetUserByToken(token string) (models.User, error) {
	pguser, err := pg.getUserByToken(token)
	if err != nil {
		return models.User{}, err
	}
	return pguser.ToModel(), err
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
	res := pg.db.Delete(user)
	if res.Error != nil {
		return errors.New("Could not delete user from the database : " + res.Error.Error())
	}
	return nil
}

// Project
func (pg *PostgreSQL) createOrUpdateProject(project pgProject) error {
	var p pgProject

	res := pg.db.Where("name = ?", project.Name).Find(&p)

	// The project doesn't exist yet
	if gorm.IsRecordNotFoundError(res.Error) {
		res := pg.db.Create(&project)
		if res.Error != nil {
			return errors.New("Could not insert project : " + res.Error.Error())
		}
		return nil
	}

	// handle other errors
	if res.Error != nil {
		return errors.New("Could not select project : " + res.Error.Error())
	}

	// update if the project was found
	res = pg.db.Model(&project).Where("name = ?", project.Name).Update(project)
	if res.Error != nil {
		return errors.New("Could not save project in the database :" + res.Error.Error())
	}

	return nil
}

func (pg *PostgreSQL) CreateOrUpdateProject(project models.Project) error {
	log.Debugf("Saving project : %s", project)

	p := pgProject{}
	p.FromModel(project)

	err := pg.createOrUpdateProject(p)
	if err != nil {
		return err
	}
	return nil
}

func (pg *PostgreSQL) getProjects() ([]pgProject, error) {
	var projects []pgProject
	res := pg.db.Find(&projects)
	if res.Error != nil {
		return nil, errors.New("Could not select projects : " + res.Error.Error())
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

	res := pg.db.Where("name = ?", projectName).Find(&project)
	if res.Error != nil {
		return project, errors.New("Could not select project : " + res.Error.Error())
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

	proj, err := pg.getProject(projectName)
	if err != nil {
		return errors.New("Could not find corresponding project for ip :" + err.Error())
	}

	ip.ProjectID = proj.ID

	res := pg.db.Create(&ip)
	if res.Error != nil {
		return errors.New("Could not save ip in the database :" + res.Error.Error())
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
	res := pg.db.
		Where("projects.name = ?", projectName).
		Find(&pgips)
	if res.Error != nil {
		return nil, errors.New("Could not get IPs : " + res.Error.Error())
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
	res := pg.db.
		Where("projects.name = ?", projectName).
		Where("ips.value = ?", ip).
		Find(&pgip)

	if gorm.IsRecordNotFoundError(res.Error) {
		return pgIP{}, nil
	}
	if res.Error != nil {
		return pgIP{}, errors.New("Could not get IP : " + res.Error.Error())
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

	res := pg.db.Create(&domain)
	if res.Error != nil {
		return errors.New("Could not save or update domain : " + res.Error.Error())
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
	res := pg.db.
		Where("projects.name = ?", projectName).
		Find(&domains)
	if res.Error != nil {
		return nil, errors.New("Could not get domains : " + res.Error.Error())
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

	res := pg.db.
		Where("projects.name = ?", projectName).
		Where("domains.name = ?", domainName).
		Find(&domain)
	if res.Error != nil {
		return pgDomain{}, errors.New("Could not get domain : " + res.Error.Error())
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

// TOFIX : doing an actual join/relation insert instead of 3 f requests
func (pg *PostgreSQL) createOrUpdatePort(projectName string, ip string, port pgPort) error {
	proj := pgProject{}
	res := pg.db.Where("name = ?", projectName).First(&proj)
	if res.Error != nil {
		return errors.New("Could not corresponding project for port : " + res.Error.Error())
	}

	pip := pgIP{}
	res = pg.db.Where("value = ?", ip).Where("project_id = ?", proj.ID).First(&pip)
	if res.Error != nil {
		return errors.New("Could not corresponding ip for port : " + res.Error.Error())
	}

	port.IPId = pip.ID
	res = pg.db.Create(&port)
	if res.Error != nil {
		return errors.New("Could not create or update port : " + res.Error.Error())
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

	res := pg.db.
		Where("ips.value = ?", ip).
		Find(&ports)
	if res.Error != nil {
		return nil, errors.New("Could not get ports : " + res.Error.Error())
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
	res := pg.db.
		Where("ips.value = ?", ip).
		Where("ports.number = ?", port).
		Find(&p)

	if res.Error != nil {
		return p, errors.New("Could not get port : " + res.Error.Error())
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
	//TOFIX : do real join / relation / whatever. Stop doing 4 request.
	proj := pgProject{}
	res := pg.db.Where("name = ?", projectName).First(&proj)
	if res.Error != nil {
		return errors.New("Could not corresponding project for port : " + res.Error.Error())
	}

	pip := pgIP{}
	res = pg.db.Where("value = ?", ip).Where("project_id = ?", proj.ID).First(&pip)
	if res.Error != nil {
		return errors.New("Could not corresponding ip for port : " + res.Error.Error())
	}

	pport := pgPort{}
	res = pg.db.Where("ip_id = ?", pip.ID).First(&pport)
	if res.Error != nil {
		return errors.New("Could not corresponding ip for port : " + res.Error.Error())
	}

	uri.PortID = pport.ID
	res = pg.db.Create(&uri)
	if res.Error != nil {
		return res.Error
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

	res := pg.db.
		Where("ips.value = ?", ip).
		Where("ports.number = ?", port).
		Find(&uris)
	if res.Error != nil {
		return uris, errors.New("Could not get URIs : " + res.Error.Error())
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

	res := pg.db.
		Where("ips.value = ?", ip).
		Where("ports.number = ?", port).
		First(&uri)

	if res.Error != nil {
		return pgURI{}, errors.New("Could not get URIs : " + res.Error.Error())
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
	res := pg.db.Create(&raw)
	if res.Error != nil {
		return errors.New("Could not insert raw : " + res.Error.Error())
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

	res := pg.db.Find(raws)
	if res.Error != nil {
		return nil, errors.New("Could not get raws : " + res.Error.Error())
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
	res := pg.db.
		Where("raws.name = ?", projectName).
		Where("raws.moduleName = ?", moduleName).
		Find(&raws)

	if res.Error != nil {
		return nil, errors.New("Could not get raw by module : " + res.Error.Error())
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
