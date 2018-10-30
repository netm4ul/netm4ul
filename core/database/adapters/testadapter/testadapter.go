package testadapter

import (
	"errors"
	"strconv"

	"github.com/netm4ul/netm4ul/core/config"
	"github.com/netm4ul/netm4ul/core/database/models"
	"github.com/netm4ul/netm4ul/tests"
)

//Test is the base struct for the "testadapter" adapter.
type Test struct {
	cfg *config.ConfigToml
}

//InitDatabase is only there to sets up the configuration of the database.
// This adapters doesn't require any connection or any other setup.
func InitDatabase(c *config.ConfigToml) *Test {
	test := Test{}
	test.cfg = c
	return &test
}

// General purpose functions

//Name return the name of the adapter. It is exported because it's called from other core packages.
func (test *Test) Name() string {
	return "TestAdapter"
}

//SetupDatabase no op. This adapter don't have any setup
func (test *Test) SetupDatabase() error {
	return nil
}

//DeleteDatabase no op.
func (test *Test) DeleteDatabase() error {
	return nil
}

//SetupAuth doesn't do anything. This adapter doesn't require any authentification setup.
func (test *Test) SetupAuth(username, password, dbname string) error {
	return nil
}

//Connect doesn't do anything. This adapter doesn't require any connection. It's there only to implement the interface.
func (test *Test) Connect(*config.ConfigToml) error {
	return nil
}

//GetUser return a models.User if the provided username correspond to the test username (stored in the /tests/values.go file).
func (test *Test) GetUser(username string) (models.User, error) {
	if tests.NormalUser.Name == username {
		return tests.NormalUser, nil
	}
	return models.User{}, models.ErrNotFound
}

//GetUserByToken return a models.User if the provided token correspond to the test token (stored in the /tests/values.go file).
func (test *Test) GetUserByToken(token string) (models.User, error) {
	if tests.NormalUser.Token == token {
		return tests.NormalUser, nil
	}
	return models.User{}, models.ErrNotFound
}

//CreateOrUpdateUser is a no-op.
func (test *Test) CreateOrUpdateUser(user models.User) error {
	exist := false
	for _, u := range tests.NormalUsers {
		if u.Name == user.Name {
			exist = true
		}
	}

	if exist {
		test.UpdateUser(user)
	} else {
		test.CreateUser(user)
	}

	return nil
}

//CreateUser is the public wrapper to create a new User in the database.
func (test *Test) CreateUser(user models.User) error {
	tests.NormalUsers = append(tests.NormalUsers, user)
	return nil
}

//UpdateUser is the public wrapper to update a new User in the database.
func (test *Test) UpdateUser(user models.User) error {
	for index, u := range tests.NormalUsers {
		if u.Name == user.Name {
			tests.NormalUsers[index] = user
		}
	}
	return nil
}

//GenerateNewToken changes the in memory value of the user token
func (test *Test) GenerateNewToken(user models.User) error {
	tests.NormalUser.Token = "Changed token"
	return nil
}

//DeleteUser delete an user or return an ErrNotFound if it doesn't exist
func (test *Test) DeleteUser(user models.User) error {
	var index int
	var u models.User

	exist := false
	for index, u = range tests.NormalUsers {
		if u.Name == user.Name {
			exist = true
			//remove an element from the slice (without preserving the order)
			// move the last element of the slice into the "index"
			// remove the last element
			tests.NormalUsers[index] = tests.NormalUsers[len(tests.NormalUsers)-1]
			tests.NormalUsers = tests.NormalUsers[:len(tests.NormalUsers)-1]
		}
	}

	if !exist {
		return models.ErrNotFound
	}
	return nil
}

// Project

//CreateOrUpdateProject is a no-op
func (test *Test) CreateOrUpdateProject(project models.Project) error {
	exist := false
	for _, p := range tests.NormalProjects {
		if p.Name == project.Name {
			exist = true
		}
	}

	if exist {
		test.UpdateProject(project)
	} else {
		test.CreateProject(project)
	}
	return nil
}

//CreateProject is the public wrapper to create a new Project in the database.
func (test *Test) CreateProject(project models.Project) error {
	tests.NormalProjects = append(tests.NormalProjects, project)
	return nil
}

//UpdateProject is the public wrapper to update a new Project in the database.
func (test *Test) UpdateProject(project models.Project) error {
	exist := false
	for index, p := range tests.NormalProjects {
		if p.Name == project.Name {
			exist = true
			tests.NormalProjects[index] = project
		}
	}

	if !exist {
		return models.ErrNotFound
	}
	return nil
}

//GetProjects returns the projects stored in the /tests/values.go file
func (test *Test) GetProjects() ([]models.Project, error) {
	return tests.NormalProjects, nil
}

//GetProject returns the project named "projectName" if it exist in the /tests/values.go file
//It returns an error if the project doesn't exist.
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
	return models.Project{}, models.ErrNotFound
}

//DeleteProject
func (test *Test) DeleteProject(project models.Project) error {
	exist := false
	for index, p := range tests.NormalProjects {
		if p.Name == project.Name {
			exist = true
			tests.NormalProjects[index] = tests.NormalProjects[len(tests.NormalProjects)-1]
			tests.NormalProjects = tests.NormalProjects[:len(tests.NormalProjects)-1]
		}
	}

	if !exist {
		return models.ErrNotFound
	}
	return nil
}

// IP

//CreateOrUpdateIP is a no-op
func (test *Test) CreateOrUpdateIP(projectName string, ip models.IP) error {
	return nil
}

//CreateIP is the public wrapper to create a new IP in the database.
func (test *Test) CreateIP(projectName string, ip models.IP) error {
	return errors.New("Not implemented yet")
}

//UpdateIP is the public wrapper to update a new IP in the database.
func (test *Test) UpdateIP(projectName string, ip models.IP) error {
	return errors.New("Not implemented yet")
}

//CreateOrUpdateIPs is a no-op
func (test *Test) CreateOrUpdateIPs(projectName string, ip []models.IP) error {
	return nil
}

//GetIPs returns the IP addresses from the /tests/values.go file
func (test *Test) GetIPs(projectName string) ([]models.IP, error) {
	return tests.NormalIPs, nil
}

//GetIP returns the IP address from the /tests/values.go file given the project name and an ip string
func (test *Test) GetIP(projectName string, ip string) (models.IP, error) {
	return tests.NormalIPs[0], nil
}

//DeleteIP TOFIX
func (test *Test) DeleteIP(ip models.IP) error {
	return errors.New("Not implemented yet")
}

// Domain

//CreateOrUpdateDomain is a no-op
func (test *Test) CreateOrUpdateDomain(projectName string, domain models.Domain) error {
	return errors.New("Not implemented yet")
}

//CreateDomain is the public wrapper to create a new Domain in the database.
func (test *Test) CreateDomain(projectName string, domain models.Domain) error {
	return errors.New("Not implemented yet")
}

//UpdateDomain is the public wrapper to update a new Domain in the database.
func (test *Test) UpdateDomain(projectName string, domain models.Domain) error {
	return errors.New("Not implemented yet")
}

//CreateOrUpdateDomains is a no-op
func (test *Test) CreateOrUpdateDomains(projectName string, domain []models.Domain) error {
	return errors.New("Not implemented yet")
}

//GetDomains return all the domains
func (test *Test) GetDomains(projectName string) ([]models.Domain, error) {
	return tests.NormalDomains, nil
}

//GetDomain TOFIX
func (test *Test) GetDomain(projectName string, domain string) (models.Domain, error) {
	return models.Domain{}, errors.New("Not implemented yet")
}

//DeleteDomain TOFIX
func (test *Test) DeleteDomain(projectName string, domain models.Domain) error {
	return errors.New("Not implemented yet")
}

// Port

//CreateOrUpdatePort is a no-op
func (test *Test) CreateOrUpdatePort(projectName string, ip string, port models.Port) error {
	return nil
}

//CreatePort is the public wrapper to create a new port in the database.
func (test *Test) CreatePort(projectName string, ip string, port models.Port) error {
	return errors.New("Not implemented yet")
}

//UpdatePort is the public wrapper to update a new port in the database.
func (test *Test) UpdatePort(projectName string, ip string, port models.Port) error {
	return errors.New("Not implemented yet")
}

//CreateOrUpdatePorts is a no-op
func (test *Test) CreateOrUpdatePorts(projectName string, ip string, port []models.Port) error {
	return nil
}

//GetPorts return multiple ports
func (test *Test) GetPorts(projectName string, ip string) ([]models.Port, error) {
	return tests.NormalPorts, nil
}

//GetPort return a port
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
	return models.Port{}, models.ErrNotFound
}

//DeletePort TOFIX
func (test *Test) DeletePort(projectName string, ip string, port models.Port) error {
	return errors.New("Not implemented yet")
}

// URI (directory and files)

//CreateOrUpdateURI is a no-op
func (test *Test) CreateOrUpdateURI(projectName string, ip string, port string, uri models.URI) error {
	return nil
}

//CreateURI is the public wrapper to create a new URI in the database.
func (test *Test) CreateURI(projectName string, ip string, port string, URI models.URI) error {
	return errors.New("Not implemented yet")
}

//UpdateURI is the public wrapper to update a new URI in the database.
func (test *Test) UpdateURI(projectName string, ip string, port string, URI models.URI) error {
	return errors.New("Not implemented yet")
}

//CreateOrUpdateURIs is a no-op
func (test *Test) CreateOrUpdateURIs(projectName string, ip string, port string, uris []models.URI) error {
	return nil
}

//GetURIs return all the URIs
func (test *Test) GetURIs(projectName string, ip string, port string) ([]models.URI, error) {
	return tests.NormalURIs, nil
}

//GetURI return one URI
func (test *Test) GetURI(projectName string, ip string, port string, uri string) (models.URI, error) {
	uris, err := test.GetURIs(projectName, ip, port)
	if err != nil {
		return models.URI{}, err
	}
	for _, u := range uris {
		if u.Name == uri {
			return u, nil
		}
	}
	return models.URI{}, models.ErrNotFound
}

// DeleteURI TOFIX
func (test *Test) DeleteURI(projectName string, ip string, port string, dir models.URI) error {
	return errors.New("Not implemented yet")
}

// Raw data

//AppendRawData is a no-op
func (test *Test) AppendRawData(projectName string, raw models.Raw) error {
	return nil
}

//GetRaws return all the raws for one project name
func (test *Test) GetRaws(projectName string) ([]models.Raw, error) {
	raws, ok := tests.NormalRaws[projectName]
	if !ok {
		return []models.Raw{}, models.ErrNotFound
	}
	return raws, nil
}

//GetRawModule return all the raws for the provided module
func (test *Test) GetRawModule(projectName string, moduleName string) (map[string][]models.Raw, error) {
	raws, err := test.GetRaws(projectName)
	if err != nil {
		return nil, err
	}

	if len(raws) == 0 {
		return nil, models.ErrNotFound
	}
	//TOFIX : return actual raw data...
	return map[string][]models.Raw{}, nil
}
