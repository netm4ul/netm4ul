package jsondb

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/netm4ul/netm4ul/core/config"
	"github.com/netm4ul/netm4ul/core/database/models"
)

/*

This database adapter is saving information inside json files.
It tries to follow the ReconJSON structure.
By default, the data are stored 2 files for each project:
	- one parsed results file
	- one raw file for storing unformatted data
One more file is used globaly for storing users data.

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

	portsFromFile, err := f.getPorts(projectName, ip)
	if err != nil {
		return err
	}

	for _, p := range portsFromFile {

		if strconv.Itoa(int(p.Number)) == port {
			p.URIs = uris
		}
	}

	return f.writePorts(projectName, ip, portsFromFile)
}

func (f *JsonDB) writePorts(projectName string, ip string, ports []jsonPort) error {
	projectFromFile, err := f.getProject(projectName)
	if err != nil {
		return err
	}

	for i, ipFromFile := range projectFromFile.IPs {
		if ipFromFile.Value == ip {
			projectFromFile.IPs[i].Ports = ports
		}
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
		return err
	}

	err = json.NewEncoder(file).Encode(project)

	if err != nil {
		return err
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

	err = json.NewEncoder(file).Encode(usersMap)
	if err != nil {
		return errors.New("Could not save json : " + err.Error())
	}

	return nil
}

func (f *JsonDB) writeRaws(file *os.File, r jsonRaws) error {
	return json.NewEncoder(file).Encode(r)
}

func (f *JsonDB) openRawFile(project, module string) (*os.File, error) {
	file, err := os.OpenFile(f.getRawPath(project, module), os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (f *JsonDB) openResultFile(project string) (*os.File, error) {
	file, err := os.OpenFile(f.getResultPath(project), os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (f *JsonDB) SetupDatabase() error {
	log.Debugf("SetupDatabase jsondb")
	//TODO : maybe create folder / files ?
	return errors.New("Not implemented yet")
}

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
func (f *JsonDB) CreateOrUpdateUser(user models.User) error {
	var u jsonUser
	u.FromModel(user)

	err := f.writeUser(u)
	if err != nil {
		return errors.New("Could not create or update user : " + err.Error())
	}
	return nil
}

/*
This function is only internaly used to get the jsonUser list.
The public GetUsers will return the generic model.User
*/
func (f *JsonDB) getUsers() ([]jsonUser, error) {
	var users map[string]jsonUser
	usersArr := []jsonUser{}

	file, err := os.Open(f.UsersPath)
	if err != nil {
		return nil, errors.New("Could not open file : " + err.Error())
	}

	err = json.NewDecoder(file).Decode(&users)
	if err == io.EOF {
		return nil, errors.New("Empty file " + f.UsersPath)
	}

	if err != nil {
		return nil, errors.New("Could not decode file : " + f.UsersPath + ", error : " + err.Error())
	}

	for _, u := range users {
		usersArr = append(usersArr, u)
	}

	return usersArr, nil
}

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

/*
GetUser is a wrapper to the getUser function. It is only used to convert a jsonUser to models.User
*/
func (f *JsonDB) GetUser(username string) (models.User, error) {
	user, err := f.getUser(username)
	if err != nil {
		return models.User{}, err
	}
	return user.ToModel(), nil
}

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

func (f *JsonDB) GenerateNewToken(user models.User) error {
	user.Token = models.GenerateNewToken()
	err := f.CreateOrUpdateUser(user)
	if err != nil {
		return errors.New("Could not generate a new token : " + err.Error())
	}
	return nil
}

func (f *JsonDB) DeleteUser(user models.User) error {
	return errors.New("Not implemented yet")
}

// Project

//CreateOrUpdateProject handle project. It will update the project if it does not exist.
func (f *JsonDB) CreateOrUpdateProject(project models.Project) error {

	projects := []jsonProject{}
	projects, err := f.getProjects()
	if err != nil {
		return err
	}

	found := false
	// copy the found projet into a jsonProject
	var jp jsonProject
	for _, p := range projects {
		if p.Name == project.Name {
			found = true
			jp = p
			break
		}
	}

	if !found {
		projects = append(projects, jp)
		return f.writeProjects(projects)
	}

	// nothing to do. Already exist
	return nil
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
		file, err := os.Open(filePath)

		if err != nil {
			log.Errorf("Could not open file : %s [err : %s]", filePath, err.Error())
			return nil, errors.New("Could not open file : " + err.Error())
		}

		err = json.NewDecoder(file).Decode(&project)
		if err == io.EOF {
			log.Errorf("Empty file %s", filePath)
			continue
		}

		if err != nil {
			log.Errorf("Could not decode file : %s [err : %s]", filePath, err.Error())
			continue
		}
		projects = append(projects, project)
	}
	return projects, err
}

//GetProjects will return all projects available. Use GetProject to select only one
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

//GetProject return only one project by its name. It use GetProjects internally
func (f *JsonDB) getProject(projectName string) (jsonProject, error) {
	projects, err := f.GetProjects()

	if err != nil {
		return jsonProject{}, errors.New("Could not get projects list : " + err.Error())
	}

	var project jsonProject

	for _, p := range projects {
		if p.Name == projectName {
			project.FromModel(p)
			return project, nil //exit early
		}
	}
	return project, nil
}

//GetProject is a wrapper around getProject.
func (f *JsonDB) GetProject(projectName string) (models.Project, error) {
	project, err := f.getProject(projectName)
	if err != nil {
		return models.Project{}, err
	}
	return project.ToModel(), nil
}

// IP

//CreateOrUpdateIP add an ip to a project if it doesn't already exist.
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
		return errors.New("Could not find this IP for project : " + projectName)
	}

	return nil
}

//CreateOrUpdateIPs is not implemented yet.
// It should be only usefull for bulk update. It might use CreateOrUpdateIP internally
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

//GetIPs is a wrapper to getIPs. It return all the IP addresses for the provided project name.
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

//GetIP returns the full data for the provided project and ip string
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

//GetIP returns the full data for the provided project and ip string
func (f *JsonDB) GetIP(projectName string, ip string) (models.IP, error) {
	ipJson, err := f.getIP(projectName, ip)
	if err != nil {
		return models.IP{}, err
	}

	return ipJson.ToModel(), nil
}

// Domain
func (f *JsonDB) CreateOrUpdateDomain(projectName string, domain models.Domain) error {
	return errors.New("Not implemented yet")
}

func (f *JsonDB) CreateOrUpdateDomains(projectName string, domain []models.Domain) error {
	return errors.New("Not implemented yet")
}

func (f *JsonDB) GetDomains(projectName string) ([]models.Domain, error) {
	return []models.Domain{}, errors.New("Not implemented yet")
}

func (f *JsonDB) GetDomain(projectName string, domain string) (models.Domain, error) {
	return models.Domain{}, errors.New("Not implemented yet")
}

// Port

//CreateOrUpdatePort create or update one port for a givent project name and ip.
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

//CreateOrUpdatePorts loop around CreateOrUpdatePort and *fail* on the first error of it.
func (f *JsonDB) CreateOrUpdatePorts(projectName string, ip string, ports []models.Port) error {
	for _, port := range ports {
		err := f.CreateOrUpdatePort(projectName, ip, port)
		if err != nil {
			return err
		}
	}
	return nil
}

//getPorts return all the port for a given project and ip
func (f *JsonDB) getPorts(projectName string, ip string) ([]jsonPort, error) {
	ipFromFile, err := f.getIP(projectName, ip)

	if err != nil {
		return nil, err
	}

	return ipFromFile.Ports, nil
}

//GetPorts is a wrapper for getPorts
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

//GetPort return only one port by it's number. Use GetPorts internally
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

//GetPort is a wrapper of getPort
func (f *JsonDB) GetPort(projectName string, ip string, port string) (models.Port, error) {
	p, err := f.getPort(projectName, ip, port)

	if err != nil {
		return models.Port{}, err
	}

	return p.ToModel(), nil
}

// URI (directory and files)

//CreateOrUpdateURI will get all the corresponding uri for a given ip, port and project combo
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

// Raw data

// AppendRawData is append only. Adds data to Raws[projectName][modules] array
func (f *JsonDB) AppendRawData(projectName string, moduleName string, data interface{}) error {

	file, err := os.OpenFile(f.getRawPath(projectName, moduleName), os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return err
	}

	// now := strconv.Itoa(int(time.Now().UnixNano()))
	now := time.Now()

	r := jsonRaws{}
	r.CreatedAt = now
	r.UpdatedAt = now
	r.Content = data

	return f.writeRaws(file, r)

}

//GetRaws return all the raws input for a project
func (f *JsonDB) getRaws(projectName string) ([]jsonRaws, error) {
	var listOfRaws []jsonRaws       // full list
	var listOfModuleRaws []jsonRaws // list by module

	files, err := filepath.Glob(f.RawGlob + projectName + "-*.json")
	if err != nil {
		return nil, err
	}

	for _, filePath := range files {

		raws := jsonRaws{}
		file, err := os.Open(filePath)
		// splitted := strings.Split(filePath, "-")
		// moduleName := strings.Replace(splitted[2], ".json", "", -1)
		err = json.NewDecoder(file).Decode(&listOfModuleRaws)

		if err != nil {
			return nil, err
		}

		listOfRaws = append(listOfRaws, raws)
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
