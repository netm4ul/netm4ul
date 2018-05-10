package jsondb

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/netm4ul/netm4ul/core/config"
	"github.com/netm4ul/netm4ul/core/database/models"
)

//JsonDB implements the models.Database interface
type JsonDB struct {
	cfg           *config.ConfigToml
	BaseDir       string
	RawPathFmt    string
	RawGlob       string
	ResultPathFmt string
	ProjectGlob   string
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

	j.cfg = c

	//ensure data folder exists
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

func (f *JsonDB) writePorts(projectName string, ip string, ports []models.Port) error {

	projectFromFile, err := f.GetProject(projectName)
	if err != nil {
		return err
	}

	for _, i := range projectFromFile.IPs {
		if i.Value == ip {
			i.Ports = ports
		}
	}

	return f.writeProject(projectFromFile)
}

func (f *JsonDB) writeProjects(projects []models.Project) error {
	for _, p := range projects {
		f, err := f.openResultFile(p.Name)
		if err != nil {
			return err
		}

		err = json.NewEncoder(f).Encode(p)

		if err != nil {
			return err
		}
	}
	return nil
}

func (f *JsonDB) writeProject(p models.Project) error {
	projects, err := f.GetProjects()
	if err != nil {
		return err
	}

	found := false
	for _, projectFromFile := range projects {
		if projectFromFile.Name == p.Name {
			found = true
			projectFromFile = p
		}
	}

	if !found {
		projects = append(projects, p)
	}

	return f.writeProjects(projects)
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

func (f *JsonDB) writeRaws(file *os.File, r map[string]interface{}) error {
	return json.NewEncoder(file).Encode(r)
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

// Project

//CreateOrUpdateProject handle project. It will update the project if it does not exist.
func (f *JsonDB) CreateOrUpdateProject(projectName string) error {
	projects, err := f.GetProjects()
	if err != nil {
		return err
	}

	found := false
	for _, p := range projects {
		if p.Name == projectName {
			found = true
		}
	}

	if !found {
		p := models.Project{Name: projectName}
		projects = append(projects, p)
		return f.writeProjects(projects)
	}

	// nothing to do. Already exist
	return nil
}

//GetProjects will return all projects available. Use GetProject to select only one
func (f *JsonDB) GetProjects() ([]models.Project, error) {
	var projects []models.Project
	var project models.Project

	files, err := filepath.Glob(f.ProjectGlob)
	if err != nil {
		return nil, err
	}
	log.Infof("Read projects files : %+v", files)

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

//GetProject return only one project by its name. It use GetProjects internally
func (f *JsonDB) GetProject(projectName string) (models.Project, error) {
	projects, err := f.GetProjects()

	if err != nil {
		return models.Project{}, errors.New("Could not get projects list : " + err.Error())
	}

	for _, p := range projects {
		if p.Name == projectName {
			return p, nil
		}
	}
	return models.Project{}, errors.New("not found")
}

// IP

//CreateOrUpdateIP add an ip to a project if it doesn't already exist.
func (f *JsonDB) CreateOrUpdateIP(projectName string, ip models.IP) error {
	project, err := f.GetProject(projectName)
	// Refactor needed
	if err != nil {
		if err.Error() == "not found" {
			// project not found, creating
			log.Infof("Creating file for project %s", projectName)
			f.openResultFile(projectName)
			project = models.Project{Name: projectName, IPs: []models.IP{ip}}

			err = f.writeProject(project)
			if err != nil {
				return errors.New("Could not save project : " + projectName)
			}
			return nil
		}
		// undefined error
		return errors.New("Could not get project " + projectName + "," + err.Error())
	}

	found := false
	for _, ipFromFile := range project.IPs {
		if ipFromFile.Value == ip.Value {
			ipFromFile = ip
			found = true
		}
	}

	if !found {
		project.IPs = append(project.IPs, ip)
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

//GetIPs returns all the IP for the provided project
func (f *JsonDB) GetIPs(projectName string) ([]models.IP, error) {
	project, err := f.GetProject(projectName)
	if err != nil {
		return []models.IP{}, err
	}

	return project.IPs, nil
}

//GetIP returns the full data for the provided project and ip string
func (f *JsonDB) GetIP(projectName string, ip string) (models.IP, error) {
	IPs, err := f.GetIPs(projectName)

	if err != nil {
		return models.IP{}, err
	}

	for _, i := range IPs {
		if i.Value == ip {
			return i, nil
		}
	}
	return models.IP{}, errors.New("not found")
}

// Port

//CreateOrUpdatePort create or update one port for a givent project name and ip.
func (f *JsonDB) CreateOrUpdatePort(projectName string, ip string, port models.Port) error {
	portsFromFile, err := f.GetPorts(projectName, ip)
	if err != nil {
		return err
	}

	found := false
	for _, p := range portsFromFile {
		if p.Number == port.Number {
			found = true
			p = port
			break
		}
	}

	if !found {
		portsFromFile = append(portsFromFile, port)
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

//GetPorts return all the port for a given project and ip
func (f *JsonDB) GetPorts(projectName string, ip string) ([]models.Port, error) {
	ipFromFile, err := f.GetIP(projectName, ip)

	if err != nil {
		return nil, err
	}

	return ipFromFile.Ports, nil
}

//GetPort return only one port by it's number. Use GetPorts internally
func (f *JsonDB) GetPort(projectName string, ip string, port string) (models.Port, error) {
	ports, err := f.GetPorts(projectName, ip)

	if err != nil {
		return models.Port{}, nil
	}
	for _, p := range ports {

		portI, err := strconv.ParseInt(port, 10, 16)
		if err != nil {
			return models.Port{}, err
		}
		if p.Number == int16(portI) {
			return p, nil
		}
	}
	return models.Port{}, errors.New("not found")
}

// Raw data

// AppendRawData is append only. Adds data to Raws[projectName][modules] array
func (f *JsonDB) AppendRawData(projectName string, moduleName string, data interface{}) error {

	file, err := os.OpenFile(f.getRawPath(projectName, moduleName), os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return err
	}

	now := strconv.Itoa(int(time.Now().UnixNano()))
	raws, err := f.GetRaws(projectName)

	if err != nil {
		//empty file, cannot parse
		if err == io.EOF {
			raws = make(models.Raws)

			modulesData := make(map[string]interface{})
			modulesData[now] = data

			raws[moduleName] = modulesData
			return f.writeRaws(file, raws[moduleName])
		}

		return errors.New("Could not get raws for project : " + err.Error())
	}

	// if the project exist
	if _, ok := raws[moduleName]; ok {
		raws[moduleName][now] = data
	} else {
		raws[moduleName] = make(map[string]interface{})
		raws[moduleName][now] = data
	}

	return f.writeRaws(file, raws[moduleName])
}

//GetRaws return all the raws input for a project
func (f *JsonDB) GetRaws(projectName string) (models.Raws, error) {
	raws := models.Raws{}

	var data map[string]interface{}

	files, err := filepath.Glob(f.RawGlob + projectName + "-*.json")
	if err != nil {
		return nil, err
	}

	for _, filePath := range files {

		file, err := os.Open(filePath)
		splitted := strings.Split(filePath, "-")
		moduleName := strings.Replace(splitted[2], ".json", "", -1)
		err = json.NewDecoder(file).Decode(&data)
		if err != nil {
			return nil, err
		}
		raws[moduleName] = data
	}
	return raws, nil
}

//GetRawModule return the raw data for one project's module
func (f *JsonDB) GetRawModule(projectName string, moduleName string) (map[string]interface{}, error) {
	res, err := f.GetRaws(projectName)
	if err != nil {
		return nil, err
	}
	log.Debugf("res : %+v", res)
	raw, ok := res[moduleName]
	if !ok {
		return nil, errors.New("not found : " + projectName + ", " + moduleName)
	}

	return raw, nil
}
