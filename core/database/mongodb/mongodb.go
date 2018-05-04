package mongodb

import (
	"fmt"
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
	m.firstConnect(c)
	return &m
}

func (mongo *MongoDB) Name() string {
	return "mongodb"
}
func (mongo *MongoDB) SetupAuth(username, password, dbname string) error {

	roles := []mgo.Role{mgo.RoleDBAdmin}
	u := mgo.User{Username: username, Password: password, Roles: roles}
	c := mongo.session.DB(dbname)

	err := c.UpsertUser(&u)
	return err
}

func (mongo *MongoDB) Connect(cfg *config.ConfigToml) {
	mongo.session = mongo.session.Clone()
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

	fmt.Println(cfg.Database.User)
	fmt.Println(cfg.Database.Database)
	s, err := mgo.DialWithInfo(mongoDBDialInfo)

	if err != nil {
		log.Fatalf("Error connecting with the database : %s", err.Error())
	}

	log.Infof("Connected to the database : %s", cfg.Database.IP)
	mongo.session = s
}

// CreateProject create a new project structure inside db
func (mongo *MongoDB) CreateProject(projectName string) {
	// mongodb will create collection on use.

	dbCollection := mongo.cfg.Database.Database
	c := mongo.session.DB(dbCollection).C("projects")

	info, err := c.Upsert(bson.M{"Name": projectName}, bson.M{"$set": bson.M{"UpdatedAt": time.Now().Unix()}})

	if mongo.cfg.Verbose && info.Updated == 1 {
		log.Infof("Info : %+v", info)
		log.Infof("Adding %s to the collections 'projects'", projectName)
	}

	if err != nil {
		log.Fatal(err)
	}
}

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

func (mongo *MongoDB) GetProjectByName(projectName string) (models.Project, error) {
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

// AppendIP is used by module to store ip data into the database
func (mongo *MongoDB) AppendIP(ip models.IP) {

	dbCollection := mongo.cfg.Database.Database
	c := mongo.session.DB(dbCollection).C("ips")

	portsArr := make([]bson.ObjectId, len(ip.Ports))
	for k := range ip.Ports {
		portsArr[k] = ip.Ports[k].ID
	}

	_, err := c.Upsert(bson.M{"Name": "ips"}, bson.M{"_id": ip.ID, "Value": ip.Value, "Ports": portsArr})
	if err != nil {
		log.Fatal("Something went wrong : ", err)
	}
}

func (mongo *MongoDB) GetIPsByProjectName(projectName string) ([]models.IP, error) {
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

func (mongo *MongoDB) GetPortsByIP(projectName string, ip string) ([]models.Port, error) {
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
	log.Debugf("project %+v", ports)

	return ports, err
}

//AppendRawData is used by module to store raw results into the database.
func (mongo *MongoDB) AppendRawData(projectName string, dataRaw interface{}) {

	data := dataRaw.(bson.M)

	dbCollection := mongo.cfg.Database.Database
	c := mongo.session.DB(dbCollection).C("raw")
	info, err := c.Upsert(bson.M{"Name": projectName}, bson.M{"$push": data})
	log.Infof("Info : %+v", info)
	if err != nil {
		log.Fatal(err)
	}
}

// AppendPorts is used by module to store ports data into the database
func (mongo *MongoDB) AppendPorts(ports []models.Port) {

	dbCollection := mongo.cfg.Database.Database
	c := mongo.session.DB(dbCollection).C("ports")

	for v := range ports {
		_, err := c.Upsert(bson.M{"Name": "ports"},
			bson.M{"_id": ports[v].ID, "Number": ports[v].Number,
				"Protocol": ports[v].Protocol,
				"Status":   ports[v].Status,
				"Banner":   ports[v].Banner})
		if err != nil {
			log.Fatal("Something went wrong (PORT) : ", err)
		}
	}
}

// UpdatePort to update a port with new information, like directories after dirb
func (mongo *MongoDB) UpdatePort(port models.Port) {

	dbCollection := mongo.cfg.Database.Database
	c := mongo.session.DB(dbCollection).C("ports")
	// Update with directories, port.Directories is []Directories and must contain IDs
	_, err := c.Upsert(bson.M{"Number": port.Number}, bson.M{"$set": bson.M{"Directories": port.Directories}})
	if err != nil {
		log.Fatal("Something went wrong (Update Port) : ", err)
	}
	//Update other stuff here
}

// UpdateProjectIPs Update IP related to a project, for now, only 1 IP
func (mongo *MongoDB) UpdateProjectIPs(projectName string, ip models.IP) {

	dbCollection := mongo.cfg.Database.Database
	c := mongo.session.DB(dbCollection).C("projects")
	_, err := c.Upsert(bson.M{"Name": projectName}, bson.M{"$set": bson.M{"UpdatedAt": time.Now().Unix(), "IPs": ip.ID}})
	if err != nil {
		log.Fatal("Something went wrong (Update Project) : ", err)
	}
}
