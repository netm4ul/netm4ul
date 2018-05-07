package files

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/netm4ul/netm4ul/core/config"
	"github.com/netm4ul/netm4ul/core/database/models"
)

var (
	//RawPathFmt defines the format's string for the raw file
	//first %s is the project name
	//seconde %s is the module name
	RawPathFmt = "./data/raw-%s-*s.json"

	//ResultPathFmt defines the format's string for all the formated result
	//%s is the project name
	ResultPathFmt = "./data/project-%s.json"
)

type FilesDB struct {
	cfg     *config.ConfigToml
	file    *os.File
	fileRaw *os.File
}

func InitDatabase(c *config.ConfigToml) *FilesDB {
	m := FilesDB{}
	m.cfg = c
	return &m
}

func GetRawPath(projectName, moduleName string) string {
	return fmt.Sprintf(RawPathFmt, projectName, moduleName)
}
func GetResultPath(projectName string) string {
	return fmt.Sprintf(ResultPathFmt, projectName)
}

func (f *FilesDB) Name() string {
	return "FilesDB"
}

func (f *FilesDB) writePorts(projectName string, ip string, ports []models.Port) error {
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

func (f *FilesDB) writeProjects(p []models.Project) error {
	return json.NewEncoder(f.file).Encode(p)
}

func (f *FilesDB) writeProject(p models.Project) error {
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

func (f *FilesDB) writeRaws(r models.Raws) error {
	return json.NewEncoder(f.fileRaw).Encode(r)
}

func (f *FilesDB) SetupAuth(username, password, dbname string) error {
	// no auth for FS ?
	return nil
}

func (f *FilesDB) Connect(c *config.ConfigToml) error {
	//dummy open file, just in case...
	file, err := os.Open(GetResultPath(c.Project.Name))
	if err != nil {
		return err
	}

	f.file = file
	return nil
}

// Project

func (f *FilesDB) CreateOrUpdateProject(projectName string) error {
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

func (f *FilesDB) GetProjects() ([]models.Project, error) {
	var projects []models.Project
	err := json.NewDecoder(f.file).Decode(projects)
	return projects, err
}

func (f *FilesDB) GetProject(projectName string) (models.Project, error) {
	projects, err := f.GetProjects()

	if err != nil {
		return models.Project{}, err
	}

	for _, p := range projects {
		if p.Name == projectName {
			return p, nil
		}
	}
	return models.Project{}, errors.New("not found")
}

// IP

func (f *FilesDB) CreateOrUpdateIP(projectName string, ip models.IP) error {
	project, err := f.GetProject(projectName)

	if err != nil {
		return err
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

	return errors.New("Could not find this IP for project : " + projectName)
}
func (f *FilesDB) CreateOrUpdateIPs(projectName string, ip []models.IP) error {
	return errors.New("Not implemented yet")
}

//GetIPs returns all the IP for the provided project
func (f *FilesDB) GetIPs(projectName string) ([]models.IP, error) {
	project, err := f.GetProject(projectName)
	if err != nil {
		return []models.IP{}, err
	}

	return project.IPs, nil
}

//GetIP returns the full data for the provided project and ip string
func (f *FilesDB) GetIP(projectName string, ip string) (models.IP, error) {
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

func (f *FilesDB) CreateOrUpdatePort(projectName string, ip string, port models.Port) error {
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
func (f *FilesDB) CreateOrUpdatePorts(projectName string, ip string, ports []models.Port) error {
	for _, port := range ports {
		err := f.CreateOrUpdatePort(projectName, ip, port)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *FilesDB) GetPorts(projectName string, ip string) ([]models.Port, error) {
	ipFromFile, err := f.GetIP(projectName, ip)

	if err != nil {
		return nil, err
	}

	return ipFromFile.Ports, nil
}

func (f *FilesDB) GetPort(projectName string, ip string, port string) (models.Port, error) {
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
func (f *FilesDB) AppendRawData(projectName string, moduleName string, data interface{}) error {
	var err error
	f.fileRaw, err = os.Open(GetResultPath(projectName))
	defer f.fileRaw.Close()

	if err != nil {
		return err
	}

	raws, err := f.GetRaws(projectName)
	if err != nil {
		return err
	}
	raws[projectName][moduleName] = append(raws[projectName][moduleName], data)

	return f.writeRaws(raws)
}

func (f *FilesDB) GetRaws(projectName string) (map[string]map[string][]interface{}, error) {
	var d models.Raws
	err := json.NewDecoder(f.fileRaw).Decode(&d)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (f *FilesDB) GetRaw(projectName string, moduleName string) ([]interface{}, error) {
	res, err := f.GetRaws(projectName)

	if err != nil {
		return nil, err
	}

	raw, ok := res[projectName][moduleName]
	if !ok {
		return nil, errors.New("not found")
	}

	return raw, nil
}
