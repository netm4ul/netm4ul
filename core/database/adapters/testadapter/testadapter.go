package testadapter

import (
	"errors"
	"strconv"

	"github.com/netm4ul/netm4ul/core/config"
	"github.com/netm4ul/netm4ul/core/database/models"
	"github.com/netm4ul/netm4ul/tests"
)

type Test struct {
	cfg *config.ConfigToml
}

func InitDatabase(c *config.ConfigToml) *Test {
	test := Test{}
	test.cfg = c
	return &test
}

// General purpose functions
func (test *Test) Name() string {
	return "TestAdapter"
}

func (test *Test) SetupAuth(username, password, dbname string) error {
	return nil
}

func (test *Test) Connect(*config.ConfigToml) error {
	return nil
}

func (test *Test) GetUser(username string) (models.User, error) {
	if tests.NormalUser.Name == username {
		return tests.NormalUser, nil
	}
	return models.User{}, errors.New("User not found")
}

func (test *Test) GetUserByToken(token string) (models.User, error) {
	if tests.NormalUser.Token == token {
		return tests.NormalUser, nil
	}
	return models.User{}, errors.New("User not found")
}

//User
func (test *Test) CreateOrUpdateUser(user models.User) error {
	return nil
}

func (test *Test) GenerateNewToken(user models.User) error {
	tests.NormalUser.Token = "Changed token"
	return nil
}

func (test *Test) DeleteUser(user models.User) error {
	return nil
}

// Project
func (test *Test) CreateOrUpdateProject(projectName models.Project) error {
	return nil
}

func (test *Test) GetProjects() ([]models.Project, error) {
	projects := []models.Project{}
	for _, p := range tests.NormalProjects {
		//removes IPs
		p.IPs = nil
		projects = append(projects, p)
	}
	return projects, nil
}

func (test *Test) GetProject(projectName string) (models.Project, error) {
	for _, p := range tests.NormalProjects {
		if p.Name == projectName {
			p.IPs = nil
			return p, nil
		}
	}
	return models.Project{}, errors.New("Could not get project " + projectName)
}

// IP
func (test *Test) CreateOrUpdateIP(projectName string, ip models.IP) error {
	return nil
}

func (test *Test) CreateOrUpdateIPs(projectName string, ip []models.IP) error {
	return nil
}

func (test *Test) GetIPs(projectName string) ([]models.IP, error) {
	return tests.NormalProject.IPs, nil
}

func (test *Test) GetIP(projectName string, ip string) (models.IP, error) {
	ips, _ := test.GetIPs(projectName)
	for _, localIp := range ips {
		if localIp.Value == ip {
			// remove Ports
			localIp.Ports = nil
			return localIp, nil
		}
	}
	return models.IP{}, errors.New("IP not found")
}

// Domain
func (test *Test) CreateOrUpdateDomain(projectName string, domain models.Domain) error {
	return errors.New("Not implemented yet")
}

func (test *Test) CreateOrUpdateDomains(projectName string, domain []models.Domain) error {
	return errors.New("Not implemented yet")
}

func (test *Test) GetDomains(projectName string) (models.Domain, error) {
	return models.Domain{}, errors.New("Not implemented yet")
}

func (test *Test) GetDomain(projectName string, domain string) (models.Domain, error) {
	return models.Domain{}, errors.New("Not implemented yet")
}

// Port
func (test *Test) CreateOrUpdatePort(projectName string, ip string, port models.Port) error {
	return nil
}

func (test *Test) CreateOrUpdatePorts(projectName string, ip string, port []models.Port) error {
	return nil
}

func (test *Test) GetPorts(projectName string, ip string) ([]models.Port, error) {
	ports := []models.Port{}
	for _, i := range tests.NormalProject.IPs {
		if i.Value == ip {
			//remove uris
			for _, p := range i.Ports {
				p.URIs = nil
				ports = append(ports, p)
			}
			return ports, nil
		}
	}
	return nil, errors.New("IP not found")
}

func (test *Test) GetPort(projectName string, ip string, port string) (models.Port, error) {
	ports, err := test.GetPorts(projectName, ip)
	if err != nil {
		return models.Port{}, err
	}

	for _, p := range ports {
		if strconv.Itoa(int(p.Number)) == port {
			p.URIs = nil
			return p, nil
		}
	}
	return models.Port{}, errors.New("Port not found")
}

// URI (directory and files)
func (test *Test) CreateOrUpdateURI(projectName string, ip string, port string, uri models.URI) error {
	return nil
}

func (test *Test) CreateOrUpdateURIs(projectName string, ip string, port string, uris []models.URI) error {
	return nil
}

func (test *Test) GetURIs(projectName string, ip string, port string) ([]models.URI, error) {
	p, err := test.GetPort(projectName, ip, port)
	if err != nil {
		return nil, err
	}

	return p.URIs, nil
}

func (test *Test) GetURI(projectName string, ip string, port string, uri string) (models.URI, error) {
	uris, err := test.GetURIs(projectName, ip, port)
	if err != nil {
		return models.URI{}, nil
	}
	for _, u := range uris {
		if u.Name == uri {
			return u, nil
		}
	}
	return models.URI{}, errors.New("Uri not found")
}

// Raw data
func (test *Test) AppendRawData(projectName string, moduleName string, data interface{}) error {
	return nil
}

func (test *Test) GetRaws(projectName string) (models.Raws, error) {
	raws, ok := tests.NormalRaws[projectName]
	if !ok {
		return models.Raws{}, errors.New("Project not found")
	}
	return raws, nil
}

func (test *Test) GetRawModule(projectName string, moduleName string) (map[string]interface{}, error) {
	raws, err := test.GetRaws(projectName)
	if err != nil {
		return nil, err
	}

	raw, ok := raws[moduleName]
	if !ok {
		return nil, errors.New("Module not found")
	}

	return raw, nil
}
