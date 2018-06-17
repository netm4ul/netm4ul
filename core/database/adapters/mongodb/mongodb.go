package mongodb

import (
	"errors"
	"time"

	"github.com/netm4ul/netm4ul/core/config"
	"github.com/netm4ul/netm4ul/core/database/models"
	log "github.com/sirupsen/logrus"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type MongoDB struct {
	cfg     *config.ConfigToml
	session *mgo.Session
}

func InitDatabase(c *config.ConfigToml) *MongoDB {
	m := MongoDB{}
	m.cfg = c
	// m.firstConnect(c)
	return &m
}

func (mongo *MongoDB) Name() string {
	return "MongoDB"
}

func (mongo *MongoDB) SetupAuth(username, password, dbname string) error {

	roles := []mgo.Role{mgo.RoleDBAdmin}
	u := mgo.User{Username: username, Password: password, Roles: roles}
	c := mongo.session.DB(dbname)

	err := c.UpsertUser(&u)
	return err
}

func (mongo *MongoDB) Connect(cfg *config.ConfigToml) error {
	mongo.session = mongo.session.Clone()
	return nil
}

// first connection to the database
func (mongo *MongoDB) firstConnect(cfg *config.ConfigToml) {
	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:    []string{cfg.Database.IP}, // array of ip (sharding & whatever), just 1 for now
		Timeout:  10 * time.Second,
		Database: cfg.Database.Database,
		Username: cfg.Database.User,
		Password: cfg.Database.Password,
	}

	log.Debugf("User : %+v", cfg.Database.User)
	log.Debugf("Datbase : %+v", cfg.Database.Database)
	s, err := mgo.DialWithInfo(mongoDBDialInfo)

	if err != nil {
		log.Fatalf("Error connecting with the database : %s", err.Error())
	}

	log.Infof("Connected to the database : %s", cfg.Database.IP)
	mongo.session = s
}

//User
func (mongo *MongoDB) CreateOrUpdateUser(user models.User) error {
	return errors.New("Not implemented yet")
}

func (mongo *MongoDB) GetUser(username string) (models.User, error) {
	return models.User{}, errors.New("Not implemented yet")
}

func (mongo *MongoDB) GetUserByToken(token string) (models.User, error) {
	return models.User{}, errors.New("Not implemented yet")
}

func (mongo *MongoDB) GenerateNewToken(user models.User) error {
	return errors.New("Not implemented yet")
}

func (mongo *MongoDB) DeleteUser(user models.User) error {
	return errors.New("Not implemented yet")
}

// CreateOrUpdateProject create a new project structure inside db
func (mongo *MongoDB) CreateOrUpdateProject(project models.Project) error {
	// mongodb will create collection on use.

	dbCollection := mongo.cfg.Database.Database
	c := mongo.session.DB(dbCollection).C("projects")

	info, err := c.Upsert(bson.M{"Name": project.Name}, bson.M{"$set": bson.M{"UpdatedAt": time.Now().Unix()}})

	if info.Updated == 1 {
		log.Debugf("Info : %+v", info)
		log.Debugf("Adding %s to the collections 'projects'", project.Name)
	}

	return err
}

//GetProjects will return all projects available. Use GetProject to select only one
func (mongo *MongoDB) GetProjects() ([]models.Project, error) {
	var project []models.Project

	dbCollection := mongo.cfg.Database.Database
	c := mongo.session.DB(dbCollection).C("projects")

	pipe := c.Pipe([]bson.M{
		{
			"$project": bson.M{
				"IPs.Ports": 0,
			},
		},
	})

	err := pipe.All(&project)

	return project, err
}

//GetProject return only one project by its name
func (mongo *MongoDB) GetProject(projectName string) (models.Project, error) {
	var project models.Project

	dbCollection := mongo.cfg.Database.Database
	c := mongo.session.DB(dbCollection).C("projects")

	// query with "join" and ignoring ports (only getting project & ips)
	pipe := c.Pipe([]bson.M{
		{
			"$match": bson.M{
				"Name": projectName,
			},
		},
		{
			"$limit": 1,
		},
		{
			"$project": bson.M{
				"IPs.Ports": 0,
			},
		},
	})

	err := pipe.One(&project)

	return project, err
}

//CreateOrUpdateIP is used by modules to store ip data into the database
//TOFIX
func (mongo *MongoDB) CreateOrUpdateIP(projectName string, ip models.IP) error {

	dbCollection := mongo.cfg.Database.Database
	c := mongo.session.DB(dbCollection).C("ips")

	portsArr := make([]bson.ObjectId, len(ip.Ports))
	for k := range ip.Ports {
		portsArr[k] = bson.ObjectIdHex(ip.Ports[k].ID)
	}

	_, err := c.Upsert(bson.M{"Name": "ips"}, bson.M{"_id": ip.ID, "Value": ip.Value, "Ports": portsArr})
	return err
}

func (mongo *MongoDB) CreateOrUpdateIPs(projectName string, ip []models.IP) error {
	return errors.New("Not implemented yet")
}

func (mongo *MongoDB) GetIPs(projectName string) ([]models.IP, error) {
	var project models.Project

	dbCollection := mongo.cfg.Database.Database
	c := mongo.session.DB(dbCollection).C("projects")

	// query with "join" and ignoring ports (only getting project & ips)
	pipe := c.Pipe([]bson.M{
		{
			"$match": bson.M{
				"Name": projectName,
			},
		},
		{
			"$lookup": bson.M{
				"from":         "ips",
				"localField":   "IPs",
				"foreignField": "_id",
				"as":           "IPs",
			},
		},
		{
			"$limit": 1,
		},
		{
			"$project": bson.M{
				"IPs.Ports": 0,
			},
		},
	})

	err := pipe.One(&project)

	return project.IPs, err
}
func (mongo *MongoDB) GetIP(projectName string, ip string) (models.IP, error) {
	return models.IP{}, errors.New("Not implemented yet")
}

// Domain
func (mongo *MongoDB) CreateOrUpdateDomain(projectName string, domain models.Domain) error {
	return errors.New("Not implemented yet")
}

func (mongo *MongoDB) CreateOrUpdateDomains(projectName string, domain []models.Domain) error {
	return errors.New("Not implemented yet")
}

func (mongo *MongoDB) GetDomains(projectName string) ([]models.Domain, error) {
	return []models.Domain{}, errors.New("Not implemented yet")
}

func (mongo *MongoDB) GetDomain(projectName string, domain string) (models.Domain, error) {
	return models.Domain{}, errors.New("Not implemented yet")
}

func (mongo *MongoDB) GetPorts(projectName string, ip string) ([]models.Port, error) {
	// var project models.Project
	var ports []models.Port

	dbCollection := mongo.cfg.Database.Database
	c := mongo.session.DB(dbCollection).C("projects")

	// query with "join" and ignoring ports (only getting project & ips)
	pipe := c.Pipe([]bson.M{
		{
			"$match": bson.M{
				"Name": projectName,
			},
		},
		{
			"$lookup": bson.M{
				"from":         "ips",
				"localField":   "IPs",
				"foreignField": "_id",
				"as":           "IPs",
			},
		},
		{
			"$lookup": bson.M{
				"from":         "ports",
				"localField":   "IPs.Ports",
				"foreignField": "_id",
				"as":           "Ports",
			},
		},
		{
			"$limit": 1,
		},
		{
			"$project": bson.M{
				"_id":       0,
				"Name":      0,
				"UpdatedAt": 0,
				"IPs":       0,
			},
		},
	})

	err := pipe.All(&ports)
	log.Debugf("====== %+v", err)
	log.Debugf("ports %+v", ports)

	return ports, err
}

func (mongo *MongoDB) GetPort(projectName string, ip string, port string) (models.Port, error) {
	return models.Port{}, errors.New("Not implemented yet")
}

// AppendPorts is used by module to store ports data into the database
func (mongo *MongoDB) CreateOrUpdatePorts(projectName string, ip string, ports []models.Port) error {

	dbCollection := mongo.cfg.Database.Database
	c := mongo.session.DB(dbCollection).C("ports")

	for v := range ports {
		_, err := c.Upsert(
			bson.M{"Name": "ports"},
			bson.M{
				"_id":      ports[v].ID,
				"Number":   ports[v].Number,
				"Protocol": ports[v].Protocol,
				"Status":   ports[v].Status,
				"Banner":   ports[v].Banner},
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// UpdatePort to update a port with new information, like URI after dirb
func (mongo *MongoDB) CreateOrUpdatePort(projectName string, ip string, port models.Port) error {
	return mongo.CreateOrUpdatePorts(projectName, ip, []models.Port{port})
}

// URI (directory and files)
func (mongo *MongoDB) CreateOrUpdateURI(projectName string, ip string, port string, uri models.URI) error {
	return errors.New("Not implemented yet")
}

func (mongo *MongoDB) CreateOrUpdateURIs(projectName string, ip string, port string, uris []models.URI) error {

	// dbCollection := mongo.cfg.Database.Database
	// c := mongo.session.DB(dbCollection).C("uris")

	return errors.New("Not implemented yet")
}

func (mongo *MongoDB) GetURIs(projectName string, ip string, port string) ([]models.URI, error) {
	var uris []models.URI

	dbCollection := mongo.cfg.Database.Database
	c := mongo.session.DB(dbCollection).C("projects")

	// query with "join" and ignoring ports (only getting project & ips)
	pipe := c.Pipe([]bson.M{
		{
			"$match": bson.M{
				"Name": projectName,
			},
		},
		{
			"$lookup": bson.M{
				"from":         "ips",
				"localField":   "IPs",
				"foreignField": "_id",
				"as":           "IPs",
			},
		},
		{
			"$lookup": bson.M{
				"from":         "ports",
				"localField":   "IPs.Ports",
				"foreignField": "_id",
				"as":           "Ports",
			},
		},
		{
			"$lookup": bson.M{
				"from":         "uri",
				"localField":   "IPs.URI",
				"foreignField": "_id",
				"as":           "URI",
			},
		},
		{
			"$project": bson.M{
				"_id":       0,
				"Name":      0,
				"UpdatedAt": 0,
				"IPs":       0,
				"Ports":     0,
			},
		},
	})

	err := pipe.All(&uris)

	log.Debugf("====== %+v", err)
	log.Debugf("URIs %+v", uris)

	if err != nil {
		return nil, err
	}

	return uris, nil
}

func (mongo *MongoDB) GetURI(projectName string, ip string, port string, dir string) (models.URI, error) {
	return models.URI{}, errors.New("Not implemented yet")
}

//AppendRawData is used by module to store raw results into the database.
func (mongo *MongoDB) AppendRawData(projectName string, moduleName string, dataRaw interface{}) error {
	data := bson.M{projectName + ".results." + moduleName: dataRaw}

	dbCollection := mongo.cfg.Database.Database
	c := mongo.session.DB(dbCollection).C("raw")
	info, err := c.Upsert(bson.M{"Name": projectName}, bson.M{"$push": data})
	log.Infof("Info : %+v", info)

	return err
}

func (mongo *MongoDB) GetRaws(projectName string) (models.Raws, error) {
	var raws models.Raws
	return raws, errors.New("Not implemented yet")
}

func (mongo *MongoDB) GetRawModule(projectName string, moduleName string) (map[string]interface{}, error) {
	return nil, errors.New("Not implemented yet")
}
