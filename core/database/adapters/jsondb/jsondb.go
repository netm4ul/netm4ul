package jsondb

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.com/netm4ul/netm4ul/core/config"
	"github.com/netm4ul/netm4ul/core/database/models"
	"github.com/netm4ul/netm4ul/core/security"
)

/*

This database adapter is saving information inside json files.
It tries to follow the ReconJSON structure.
By default, the data are stored 2 files for each project:
	- one parsed results file
	- one raw file for storing unformatted data
One more file is used globally for storing users data.

*/

//JsonDB implements the models.Database interface
type JsonDB struct {
	cfg           *config.ConfigToml
	BaseDir       string
	RawPathFmt    string
	RawGlob       string
	ResultPathFmt string
	ProjectGlob   string
	UsersPath     string
}

//InitDatabase check if data dir is created and return the JsonDB struct
func InitDatabase(c *config.ConfigToml) *JsonDB {

	j := JsonDB{}

	//BaseDir is the data folder
	j.BaseDir = "./data"

	//RawPathFmt defines the format's string for the raw file
	//first %s is the project name
	//second %s is the module name
	j.RawPathFmt = j.BaseDir + "/raw-%s-%s.json"
	// RawGlob is the prefix glob
	j.RawGlob = j.BaseDir + "/raw-"

	//ResultPathFmt defines the format's string for all the formated result
	//%s is the project name
	j.ResultPathFmt = j.BaseDir + "/project-%s.json"
	//ProjectGlob defines the glob for project files
	j.ProjectGlob = j.BaseDir + "/project-*"
	//UserPath defines the path for API user info
	j.UsersPath = j.BaseDir + "/users.json"
	j.cfg = c

	//ensure data folder exists
	//TOFIX : call that only on the root executable (tests creates ./data in the jsondb folder...)
	if _, err := os.Stat(j.BaseDir); os.IsNotExist(err) {
		os.Mkdir(j.BaseDir, 0755)
	}

	return &j
}

func (f *JsonDB) getRawPath(projectName, moduleName string) string {
	return fmt.Sprintf(f.RawPathFmt, projectName, moduleName)
}
func (f *JsonDB) getResultPath(projectName string) string {
	return fmt.Sprintf(f.ResultPathFmt, projectName)
}

//Name return the adapter name
func (f *JsonDB) Name() string {
	return "JsonDB"
}

//TOFIX
//these writes should all have locks
func (f *JsonDB) writeURIs(projectName string, ip string, port string, uris []jsonURI) error {
	if projectName == "" {
		return errors.New("empty project name (writeURIs)")
	}
	if ip == "" {
		return errors.New("empty ip (writeURIs)")
	}
	if port == "" {
		return errors.New("empty port (writeURIs)")
	}

	portsFromFile, err := f.getPorts(projectName, ip)
	if err != nil {
		return err
	}
	found := 0
	var i int
	for i, p := range portsFromFile {

		if strconv.Itoa(int(p.Number)) == port {
			found = i
			p.URIs = uris
		}
	}

	if found > 0 {
		portsFromFile[i].URIs = append(portsFromFile[i].URIs, uris...)
	}

	return f.writePorts(projectName, ip, portsFromFile)
}

func (f *JsonDB) writePorts(projectName string, ip string, ports []jsonPort) error {
	if projectName == "" {
		return errors.New("empty project name (writePorts)")
	}
	if ip == "" {
		return errors.New("empty ip (writePorts)")
	}

	projectFromFile, err := f.getProject(projectName)
	if err != nil {
		return err
	}

	found := 0
	for i, ipFromFile := range projectFromFile.IPs {
		if ipFromFile.Value == ip {
			ipFromFile.Ports = ports
			found = i
		}
	}

	// project not found
	if found == 0 {
		projectFromFile.IPs[found].Ports = append(projectFromFile.IPs[found].Ports, ports...)
	}

	return f.writeProjects([]jsonProject{projectFromFile})
}

/*
writeProjects will write every project in its own project file
*/
func (f *JsonDB) writeProjects(projects []jsonProject) error {
	for _, p := range projects {
		f.writeProject(p)
	}
	return nil
}

/*

 */
func (f *JsonDB) writeProject(project jsonProject) error {

	file, err := f.openResultFile(project.Name)
	if err != nil {
		return errors.New("Could not open result files : " + err.Error())
	}

	err = json.NewEncoder(file).Encode(project)
	if err != nil {
		return errors.New("Could not encode this project : " + err.Error())
	}

	return nil
}

func (f *JsonDB) writeUser(user jsonUser) error {
	users, err := f.getUsers()
	if err != nil {
		return errors.New("Could not write user : " + err.Error())
	}

	var usersMap map[string]jsonUser
	usersMap = make(map[string]jsonUser, 0)

	// add or replace user !
	for _, u := range users {
		usersMap[u.Name] = u
	}
	usersMap[user.Name] = user

	file, err := os.OpenFile(f.UsersPath, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return errors.New("Could not open file : " + err.Error())
	}
	defer file.Close()

	err = json.NewEncoder(file).Encode(usersMap)
	if err != nil {
		return errors.New("Could not save json : " + err.Error())
	}

	return nil
}

// append only
func (f *JsonDB) writeRaw(file *os.File, r jsonRaws) error {

	log.Debugf("r : %+v", r)

	if r.ProjectName == "" {
		return errors.New("empty project (writeRaw)")
	}

	raws, err := f.getRaws(r.ProjectName)
	if err != nil {
		return errors.New("Could not getRaws for the project name : " + err.Error())
	}

	raws = append(raws, r)
	err = json.NewEncoder(file).Encode(raws)
	if err != nil {
		return err
	}

	file.Close()
	return nil
}

func (f *JsonDB) openRawFile(project, module string) (*os.File, error) {
	if project == "" {
		return nil, errors.New("empty project name (openRawFile)")
	}
	if module == "" {
		return nil, errors.New("empty module name (openRawFile)")
	}

	file, err := os.OpenFile(f.getRawPath(project, module), os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return file, nil
}

func (f *JsonDB) openResultFile(project string) (*os.File, error) {
	if project == "" {
		return nil, errors.New("empty project name")
	}

	file, err := os.OpenFile(f.getResultPath(project), os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return file, nil
}

//SetupDatabase TODO
func (f *JsonDB) SetupDatabase() error {
	log.Debugf("SetupDatabase jsondb")
	//TODO : maybe create folder / files ?
	return errors.New("Not implemented yet")
}

//DeleteDatabase TODO
func (f *JsonDB) DeleteDatabase() error {
	return errors.New("Not implemented yet")
}

//SetupAuth is not used in this adapters. Filesystem permission are used for that
func (f *JsonDB) SetupAuth(username, password, dbname string) error {
	// no auth for FS ?
	return nil
}

//Connect is only there to fully implement the models.Database interface. Always return nil
func (f *JsonDB) Connect(c *config.ConfigToml) error {
	return nil
}

//User

//CreateOrUpdateUser create or update a new user
func (f *JsonDB) CreateOrUpdateUser(user models.User) error {
	var u jsonUser
	u.FromModel(user)

	err := f.writeUser(u)
	if err != nil {
		return errors.New("Could not create or update user : " + err.Error())
	}
	return nil
}

//CreateUser is the public wrapper to create a new User in the database.
func (f *JsonDB) CreateUser(user models.User) error {
	return errors.New("Not implemented yet")
}

//UpdateUser is the public wrapper to update a new User in the database.
func (f *JsonDB) UpdateUser(user models.User) error {
	return errors.New("Not implemented yet")
}

// This function is only internaly used to get the jsonUser list.
// The public GetUsers will return the generic model.User
func (f *JsonDB) getUsers() ([]jsonUser, error) {
	var users map[string]jsonUser
	usersArr := []jsonUser{}

	file, err := os.OpenFile(f.UsersPath, os.O_RDONLY|os.O_CREATE, 0755)
	if err != nil {
		return nil, errors.New("Could not open file : " + err.Error())
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&users)
	if err == io.EOF {
		return usersArr, nil
	}

	if err != nil {
		return nil, errors.New("Could not decode file : " + f.UsersPath + ", error : " + err.Error())
	}

	for _, u := range users {
		usersArr = append(usersArr, u)
	}

	return usersArr, nil
}

//GetUsers returns all the users stored in the json database
func (f *JsonDB) GetUsers() ([]models.User, error) {

	usersModel := []models.User{}
	users, err := f.getUsers()

	if err != nil {
		return nil, err
	}

	for _, user := range users {
		usersModel = append(usersModel, user.ToModel())
	}

	return usersModel, nil

}

func (f *JsonDB) getUser(username string) (jsonUser, error) {

	users, err := f.getUsers()
	if err != nil {
		return jsonUser{}, err
	}

	for _, user := range users {
		if user.Name == username {
			return user, nil
		}
	}

	return jsonUser{}, nil
}

// GetUser is a wrapper to the getUser function. It is only used to convert a jsonUser to models.User
func (f *JsonDB) GetUser(username string) (models.User, error) {
	user, err := f.getUser(username)
	if err != nil {
		return models.User{}, err
	}
	return user.ToModel(), nil
}

// GetUserByToken will return the user information by it's token
func (f *JsonDB) GetUserByToken(token string) (models.User, error) {

	users, err := f.getUsers()
	if err != nil {
		return models.User{}, err
	}
	for _, user := range users {
		if user.Token == token {
			return user.ToModel(), nil
		}
	}

	return models.User{}, nil
}

// GenerateNewToken generates a new token and save it in the database.
// It uses the function GenerateNewToken provided by the `models` class
func (f *JsonDB) GenerateNewToken(user models.User) error {
	user.Token = security.GenerateNewToken()
	err := f.CreateOrUpdateUser(user)
	if err != nil {
		return errors.New("Could not generate a new token : " + err.Error())
	}
	return nil
}

// DeleteUser remove the user from the database (using its ID)
// TODO
func (f *JsonDB) DeleteUser(user models.User) error {
	return errors.New("Not implemented yet")
}

// Project

// CreateOrUpdateProject handle project. It will update the project if it does not exist.
func (f *JsonDB) CreateOrUpdateProject(project models.Project) error {
	log.Debug("Create Project !")
	jp := jsonProject{}
	jp.FromModel(project)
	return f.createOrUpdateProject(jp)
}

//CreateProject is the public wrapper to create a new Project in the database.
func (f *JsonDB) CreateProject(project models.Project) error {
	return errors.New("Not implemented yet")
}

//UpdateProject is the public wrapper to update a new Project in the database.
func (f *JsonDB) UpdateProject(project models.Project) error {
	return errors.New("Not implemented yet")
}

func (f *JsonDB) createOrUpdateProject(project jsonProject) error {

	projects, err := f.getProjects()
	if err != nil {
		return err
	}

	found := false
	for _, p := range projects {
		if p.Name == project.Name {
			found = true
			p = project
			break
		}
	}

	if !found {
		projects = append(projects, project)
	}

	return f.writeProjects(projects)
}

func (f *JsonDB) getProjects() ([]jsonProject, error) {
	var project jsonProject

	projects := []jsonProject{}

	files, err := filepath.Glob(f.ProjectGlob)
	if err != nil {
		return nil, err
	}
	log.Debugf("Read projects files : %+v", files)

	for _, filePath := range files {

		file, err := os.OpenFile(filePath, os.O_RDONLY|os.O_CREATE, 0755)
		if err != nil {
			log.Errorf("Could not open file : %s [err : %s]", filePath, err.Error())
			return nil, errors.New("Could not open file : " + err.Error())
		}
		defer file.Close()

		err = json.NewDecoder(file).Decode(&project)

		if err != nil {
			// skip empty file error
			if err == io.EOF {
				log.Infof("Empty file %s", filePath)
				continue
			}
			return nil, errors.New("Could not decode file : " + err.Error())
		}

		projects = append(projects, project)
	}

	return projects, nil
}

// GetProjects will return all projects available. Use GetProject to select only one
func (f *JsonDB) GetProjects() ([]models.Project, error) {
	projectsModel := []models.Project{}
	projects, err := f.getProjects()

	if err != nil {
		return nil, err
	}

	for _, p := range projects {
		projectsModel = append(projectsModel, p.ToModel())
	}
	return projectsModel, nil
}

// GetProject return only one project by its name. It use GetProjects internally
func (f *JsonDB) getProject(projectName string) (jsonProject, error) {
	projects, err := f.getProjects()

	if err != nil {
		return jsonProject{}, errors.New("Could not get projects list : " + err.Error())
	}

	for _, p := range projects {
		if p.Name == projectName {
			return p, nil //exit early
		}
	}
	return jsonProject{}, nil
}

// GetProject is a wrapper around getProject.
func (f *JsonDB) GetProject(projectName string) (models.Project, error) {
	project, err := f.getProject(projectName)
	if err != nil {
		return models.Project{}, err
	}
	return project.ToModel(), nil
}

// DeleteProject deletes the given project
// TODO
func (f *JsonDB) DeleteProject(project models.Project) error {
	return errors.New("Not implemented yet")
}

// IP

// CreateOrUpdateIP add an ip to a project if it doesn't already exist.
func (f *JsonDB) CreateOrUpdateIP(projectName string, ip models.IP) error {
	project, err := f.getProject(projectName)
	// Refactor needed
	if err != nil {
		if err.Error() == "not found" {
			// project not found, creating
			log.Infof("Creating file for project %s", projectName)
			f.openResultFile(projectName)

			p := jsonProject{}
			p.Name = projectName

			ipJson := jsonIP{}
			ipJson.FromModel(ip)

			p.IPs = []jsonIP{ipJson}

			err = f.writeProject(p)
			if err != nil {
				return errors.New("Could not save project : " + projectName)
			}
			return nil
		}
		// undefined error
		return errors.New("Could not get project " + projectName + "," + err.Error())
	}

	found := false
	ipj := jsonIP{}
	ipj.FromModel(ip)

	for _, ipFromFile := range project.IPs {
		if ipFromFile.Value == ip.Value {
			ipFromFile = ipj
			found = true
			break
		}
	}

	if !found {
		project.IPs = append(project.IPs, ipj)
	}

	err = f.writeProject(project)
	if err != nil {
		return errors.New("Could not write project " + projectName + " : " + err.Error())
	}

	return nil
}

//CreateIP is the public wrapper to create a new IP in the database.
func (f *JsonDB) CreateIP(project string, ip models.IP) error {
	return errors.New("Not implemented yet")
}

//UpdateIP is the public wrapper to update a new IP in the database.
func (f *JsonDB) UpdateIP(project string, ip models.IP) error {
	return errors.New("Not implemented yet")
}

// CreateOrUpdateIPs is not implemented yet.
// It should be only useful for bulk update. It might use CreateOrUpdateIP internally
func (f *JsonDB) CreateOrUpdateIPs(projectName string, ip []models.IP) error {
	return errors.New("Not implemented yet")
}

func (f *JsonDB) getIPs(projectName string) ([]jsonIP, error) {
	project, err := f.getProject(projectName)
	if err != nil {
		return nil, err
	}

	return project.IPs, nil
}

// GetIPs is a wrapper to getIPs. It return all the IP addresses for the provided project name.
func (f *JsonDB) GetIPs(projectName string) ([]models.IP, error) {
	ipsModel := []models.IP{}

	ips, err := f.getIPs(projectName)
	if err != nil {
		return nil, err
	}
	for _, ip := range ips {
		ipsModel = append(ipsModel, ip.ToModel())
	}

	return ipsModel, nil
}

// GetIP returns the full data for the provided project and ip string
func (f *JsonDB) getIP(projectName string, ip string) (jsonIP, error) {
	IPs, err := f.getIPs(projectName)

	if err != nil {
		return jsonIP{}, err
	}

	// return only the selected ip
	for _, i := range IPs {
		if i.Value == ip {
			return i, nil
		}
	}
	return jsonIP{}, nil
}

// GetIP returns the full data for the provided project and ip string
func (f *JsonDB) GetIP(projectName string, ip string) (models.IP, error) {
	ipJson, err := f.getIP(projectName, ip)
	if err != nil {
		return models.IP{}, err
	}

	return ipJson.ToModel(), nil
}

// DeleteIP delete the provided IP from the database
//TODO
func (f *JsonDB) DeleteIP(ip models.IP) error {
	return errors.New("Not implemented yet")
}

// Domain

// CreateOrUpdateDomain creates or updates a (sub)domain name.
// TODO
func (f *JsonDB) CreateOrUpdateDomain(projectName string, domain models.Domain) error {
	return errors.New("Not implemented yet")
}

//CreateDomain is the public wrapper to create a new Domain in the database.
func (f *JsonDB) CreateDomain(projectName string, domain models.Domain) error {
	return errors.New("Not implemented yet")
}

//UpdateDomain is the public wrapper to update a new Domain in the database.
func (f *JsonDB) UpdateDomain(projectName string, domain models.Domain) error {
	return errors.New("Not implemented yet")
}

// CreateOrUpdateDomains creates or updates multiple (sub)domain name.
// It should be used instead of CreateOrUpdateDomain for bulk insert
// TODO
func (f *JsonDB) CreateOrUpdateDomains(projectName string, domain []models.Domain) error {
	return errors.New("Not implemented yet")
}

// GetDomains return all the domains for a given project
// TODO
func (f *JsonDB) GetDomains(projectName string) ([]models.Domain, error) {
	return []models.Domain{}, errors.New("Not implemented yet")
}

// GetDomain return informaiton about the given domain and project combo
// TODO
func (f *JsonDB) GetDomain(projectName string, domain string) (models.Domain, error) {
	return models.Domain{}, errors.New("Not implemented yet")
}

//DeleteDomain remove a given domain from the database
//TODO
func (f *JsonDB) DeleteDomain(projectName string, domain models.Domain) error {
	return errors.New("Not implemented yet")
}

// Port

// CreateOrUpdatePort create or update one port for a givent project name and ip.
func (f *JsonDB) CreateOrUpdatePort(projectName string, ip string, portModel models.Port) error {
	portsFromFile, err := f.getPorts(projectName, ip)
	if err != nil {
		return err
	}

	portJson := jsonPort{}
	portJson.FromModel(portModel)

	found := false
	for _, p := range portsFromFile {
		if p.Number == portModel.Number {
			found = true

			// add missings fields if it already exist !
			portJson.PortType = p.PortType
			portJson.ID = p.ID
			p = portJson
			break
		}
	}

	if !found {
		portsFromFile = append(portsFromFile, portJson)
	}

	return f.writePorts(projectName, ip, portsFromFile)
}

//CreatePort is the public wrapper to create a new Port in the database.
func (f *JsonDB) CreatePort(projectName string, ip string, port models.Port) error {
	return errors.New("Not implemented yet")
}

//UpdatePort is the public wrapper to update a new Port in the database.
func (f *JsonDB) UpdatePort(projectName string, ip string, port models.Port) error {
	return errors.New("Not implemented yet")
}

// CreateOrUpdatePorts loop around CreateOrUpdatePort and *fail* on the first error of it.
func (f *JsonDB) CreateOrUpdatePorts(projectName string, ip string, ports []models.Port) error {
	for _, port := range ports {
		err := f.CreateOrUpdatePort(projectName, ip, port)
		if err != nil {
			return err
		}
	}
	return nil
}

// getPorts return all the port for a given project and ip
func (f *JsonDB) getPorts(projectName string, ip string) ([]jsonPort, error) {
	ipFromFile, err := f.getIP(projectName, ip)

	if err != nil {
		return nil, err
	}

	return ipFromFile.Ports, nil
}

// GetPorts is a wrapper for getPorts
func (f *JsonDB) GetPorts(projectName string, ip string) ([]models.Port, error) {
	portsModel := []models.Port{}
	ports, err := f.getPorts(projectName, ip)
	if err != nil {
		return nil, err
	}

	for _, p := range ports {
		portsModel = append(portsModel, p.ToModel())
	}
	return portsModel, nil
}

// GetPort return only one port by it's number. Use GetPorts internally
func (f *JsonDB) getPort(projectName string, ip string, port string) (jsonPort, error) {
	ports, err := f.getPorts(projectName, ip)

	if err != nil {
		return jsonPort{}, err
	}

	for _, p := range ports {

		portI, err := strconv.ParseInt(port, 10, 16)
		if err != nil {
			return jsonPort{}, err
		}
		if p.Number == int16(portI) {
			return p, nil
		}
	}
	return jsonPort{}, nil
}

// GetPort is a wrapper of getPort
func (f *JsonDB) GetPort(projectName string, ip string, port string) (models.Port, error) {
	p, err := f.getPort(projectName, ip, port)

	if err != nil {
		return models.Port{}, err
	}

	return p.ToModel(), nil
}

//DeletePort remove a given port from the database. It will need a projectName and an IP.
func (f *JsonDB) DeletePort(projectName string, ip string, port models.Port) error {
	return errors.New("Not implemented yet")
}

// URI (directory and files)

// CreateOrUpdateURI will get all the corresponding uri for a given ip, port and project combo
func (f *JsonDB) CreateOrUpdateURI(projectName string, ip string, port string, uri models.URI) error {
	urisFromFile, err := f.getURIs(projectName, ip, port)
	if err != nil {
		return err
	}

	uriJson := jsonURI{}
	uriJson.FromModel(uri)

	found := false
	for i, u := range urisFromFile {
		if uri.Name == u.Name {
			found = true

			//add missing field
			uriJson.ID = u.ID
			uriJson.Port = u.Port
			urisFromFile[i] = uriJson
			break
		}
	}

	if !found {
		urisFromFile = append(urisFromFile, uriJson)
	}
	return f.writeURIs(projectName, ip, port, urisFromFile)
}

//CreateURI is the public wrapper to create a new URI in the database.
func (f *JsonDB) CreateURI(projectName string, ip string, port string, uri models.URI) error {
	return errors.New("Not implemented yet")
}

//UpdateURI is the public wrapper to update a new URI in the database.
func (f *JsonDB) UpdateURI(projectName string, ip string, port string, uri models.URI) error {
	return errors.New("Not implemented yet")
}

// CreateOrUpdateURIs create or update a given URI
func (f *JsonDB) CreateOrUpdateURIs(projectName string, ip string, port string, uris []models.URI) error {
	for _, uri := range uris {
		err := f.CreateOrUpdateURI(projectName, ip, port, uri)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *JsonDB) getURIs(projectName string, ip string, port string) ([]jsonURI, error) {
	p, err := f.getPort(projectName, ip, port)
	if err != nil {
		return nil, errors.New("Could not get URI : " + err.Error())
	}
	return p.URIs, nil
}

// GetURIs returns all the URL from a project, IP, port combo.
func (f *JsonDB) GetURIs(projectName string, ip string, port string) ([]models.URI, error) {
	uris := []models.URI{}
	p, err := f.getPort(projectName, ip, port)
	if err != nil {
		return nil, errors.New("Could not get URI : " + err.Error())
	}

	//convert to models.URI
	for _, uriJson := range p.URIs {
		uris = append(uris, uriJson.ToModel())
	}
	return uris, nil
}

// GetURI returns information about the URL from a project, IP, port combo.
func (f *JsonDB) GetURI(projectName string, ip string, port string, uri string) (models.URI, error) {
	uris, err := f.getURIs(projectName, ip, port)

	if err != nil {
		return models.URI{}, err
	}

	for _, u := range uris {
		if u.Name == uri {
			return u.ToModel(), nil
		}
	}
	return models.URI{}, nil
}

func (f *JsonDB) DeleteURI(projectName string, ip string, port string, dir models.URI) error {
	return errors.New("Not implemented yet")
}

// Raw data

// AppendRawData is append only. Adds data to Raws[projectName][modules] array
func (f *JsonDB) AppendRawData(projectName string, raw models.Raw) error {

	file, err := os.OpenFile(f.getRawPath(projectName, raw.ModuleName), os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return errors.New("Could not append raw data :" + err.Error())
	}

	r := jsonRaws{}
	r.FromModel(raw)
	r.ProjectName = projectName

	err = f.writeRaw(file, r)
	if err != nil {
		return errors.New("Could not write raw data : " + err.Error())
	}

	return nil

}

//GetRaws return all the raws input for a project
func (f *JsonDB) getRaws(projectName string) ([]jsonRaws, error) {
	listOfRaws := []jsonRaws{}       // full list
	listOfModuleRaws := []jsonRaws{} // list by module

	files, err := filepath.Glob(f.RawGlob + projectName + "-*.json")
	if err != nil {
		return nil, err
	}

	for _, filePath := range files {

		file, err := os.OpenFile(filePath, os.O_RDONLY|os.O_CREATE, 0755)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		err = json.NewDecoder(file).Decode(&listOfModuleRaws)
		if err == io.EOF {
			continue
		}
		if err != nil {
			return nil, err
		}

		listOfRaws = append(listOfRaws, listOfModuleRaws...)
	}

	return listOfRaws, nil
}

//GetRaws return all the raws input for a project
func (f *JsonDB) GetRaws(projectName string) ([]models.Raw, error) {
	listJsonRaws, err := f.getRaws(projectName)
	if err != nil {
		return nil, err
	}
	listOfRaws := []models.Raw{}
	for _, raws := range listJsonRaws {
		listOfRaws = append(listOfRaws, raws.ToModel())
	}
	return listOfRaws, nil
}

/*
GetRawModule return the raw data for one module of one project
The result is a map of list of raw. The map key is the name of the module.
Every Raw is a unique run of a program.
It also include it's module name, but for the sake of searching through it, we mapped the module name as it's key.
*/
func (f *JsonDB) GetRawModule(projectName string, moduleName string) (map[string][]models.Raw, error) {
	raws, err := f.GetRaws(projectName)
	if err != nil {
		return nil, err
	}

	var res map[string][]models.Raw
	res = make(map[string][]models.Raw)

	for _, r := range raws {
		res[r.ModuleName] = append(res[r.ModuleName], r)
	}

	return res, nil
}
