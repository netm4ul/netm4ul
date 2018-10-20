package testadapter

import (
	"bytes"
	"encoding/gob"
	"errors"
	"strconv"

	"github.com/netm4ul/netm4ul/core/config"
	"github.com/netm4ul/netm4ul/core/database/models"
	"github.com/netm4ul/netm4ul/tests"
)

func Clone(a, b interface{}) error {
	buff := new(bytes.Buffer)
	enc := gob.NewEncoder(buff)
	dec := gob.NewDecoder(buff)
	err := enc.Encode(a)
	if err != nil {
		return err
	}
	err = dec.Decode(b)
	if err != nil {
		return err
	}
	return nil
}

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

func (test *Test) SetupDatabase() error {
	return errors.New("Not implemented yet")
}

func (test *Test) DeleteDatabase() error {
	return errors.New("Not implemented yet")
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
	return tests.NormalProjects, nil
}

func (test *Test) GetProject(projectName string) (models.Project, error) {
	ps, err := test.GetProjects()
	var projects []models.Project
	projects = append([]models.Project{}, ps...)

	if err != nil {
		return models.Project{}, errors.New("Could not get projects" + err.Error())
	}

	for _, p := range projects {
		if p.Name == projectName {
			return p, nil
		}
	}
	return models.Project{}, errors.New("Could not get project " + projectName)
}

func (test *Test) DeleteProject(project models.Project) error {
	return errors.New("Not implemented yet")
}

// IP
func (test *Test) CreateOrUpdateIP(projectName string, ip models.IP) error {
	return nil
}

func (test *Test) CreateOrUpdateIPs(projectName string, ip []models.IP) error {
	return nil
}

func (test *Test) GetIPs(projectName string) ([]models.IP, error) {
	return tests.NormalIPs, nil
}

func (test *Test) GetIP(projectName string, ip string) (models.IP, error) {
	return tests.NormalIPs[0], nil
}

func (test *Test) DeleteIP(ip models.IP) error {
	return errors.New("Not implemented yet")
}

// Domain
func (test *Test) CreateOrUpdateDomain(projectName string, domain models.Domain) error {
	return errors.New("Not implemented yet")
}

func (test *Test) CreateOrUpdateDomains(projectName string, domain []models.Domain) error {
	return errors.New("Not implemented yet")
}

func (test *Test) GetDomains(projectName string) ([]models.Domain, error) {
	return tests.NormalDomains, nil
}

func (test *Test) GetDomain(projectName string, domain string) (models.Domain, error) {
	return models.Domain{}, errors.New("Not implemented yet")
}

func (test *Test) DeleteDomain(projectName string, domain models.Domain) error {
	return errors.New("Not implemented yet")
}

// Port
func (test *Test) CreateOrUpdatePort(projectName string, ip string, port models.Port) error {
	return nil
}

func (test *Test) CreateOrUpdatePorts(projectName string, ip string, port []models.Port) error {
	return nil
}

func (test *Test) GetPorts(projectName string, ip string) ([]models.Port, error) {
	return tests.NormalPorts, nil
}

func (test *Test) GetPort(projectName string, ip string, port string) (models.Port, error) {
	ports, err := test.GetPorts(projectName, ip)
	if err != nil {
		return models.Port{}, err
	}

	for _, p := range ports {
		if strconv.Itoa(int(p.Number)) == port {
			return p, nil
		}
	}
	return models.Port{}, errors.New("Port not found")
}

func (test *Test) DeletePort(projectName string, ip string, port models.Port) error {
	return errors.New("Not implemented yet")
}

// URI (directory and files)
func (test *Test) CreateOrUpdateURI(projectName string, ip string, port string, uri models.URI) error {
	return nil
}

func (test *Test) CreateOrUpdateURIs(projectName string, ip string, port string, uris []models.URI) error {
	return nil
}

func (test *Test) GetURIs(projectName string, ip string, port string) ([]models.URI, error) {

	return tests.NormalURIs, nil
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

func (test *Test) DeleteURI(projectName string, ip string, port string, dir models.URI) error {
	return errors.New("Not implemented yet")
}

// Raw data
func (test *Test) AppendRawData(projectName string, raw models.Raw) error {
	return nil
}

func (test *Test) GetRaws(projectName string) ([]models.Raw, error) {
	raws, ok := tests.NormalRaws[projectName]
	if !ok {
		return []models.Raw{}, errors.New("Project not found")
	}
	return raws, nil
}

func (test *Test) GetRawModule(projectName string, moduleName string) (map[string][]models.Raw, error) {
	raws, err := test.GetRaws(projectName)
	if err != nil {
		return nil, err
	}

	if len(raws) == 0 {
		return nil, errors.New("Module not found")
	}
	//TOFIX : return actual raw data...
	return map[string][]models.Raw{}, nil
}
