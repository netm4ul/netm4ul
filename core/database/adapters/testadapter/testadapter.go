package testadapter

// import (
// 	"bytes"
// 	"encoding/gob"
// 	"errors"
// 	"fmt"
// 	"strconv"

// 	"github.com/netm4ul/netm4ul/core/config"
// 	"github.com/netm4ul/netm4ul/core/database/models"
// 	"github.com/netm4ul/netm4ul/tests"
// )

// func Clone(a, b interface{}) error {
// 	buff := new(bytes.Buffer)
// 	enc := gob.NewEncoder(buff)
// 	dec := gob.NewDecoder(buff)
// 	err := enc.Encode(a)
// 	if err != nil {
// 		return err
// 	}
// 	err = dec.Decode(b)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// type Test struct {
// 	cfg *config.ConfigToml
// }

// func InitDatabase(c *config.ConfigToml) *Test {
// 	test := Test{}
// 	test.cfg = c
// 	return &test
// }

// // General purpose functions
// func (test *Test) Name() string {
// 	return "TestAdapter"
// }

// func (test *Test) SetupDatabase() error {
// 	return errors.New("Not implemented yet")
// }

// func (test *Test) DeleteDatabase() error {
// 	return errors.New("Not implemented yet")
// }

// func (test *Test) SetupAuth(username, password, dbname string) error {
// 	return nil
// }

// func (test *Test) Connect(*config.ConfigToml) error {
// 	return nil
// }

// func (test *Test) GetUser(username string) (models.User, error) {
// 	if tests.NormalUser.Name == username {
// 		return tests.NormalUser, nil
// 	}
// 	return models.User{}, errors.New("User not found")
// }

// func (test *Test) GetUserByToken(token string) (models.User, error) {
// 	if tests.NormalUser.Token == token {
// 		return tests.NormalUser, nil
// 	}
// 	return models.User{}, errors.New("User not found")
// }

// //User
// func (test *Test) CreateOrUpdateUser(user models.User) error {
// 	return nil
// }

// func (test *Test) GenerateNewToken(user models.User) error {
// 	tests.NormalUser.Token = "Changed token"
// 	return nil
// }

// func (test *Test) DeleteUser(user models.User) error {
// 	return nil
// }

// // Project
// func (test *Test) CreateOrUpdateProject(projectName models.Project) error {
// 	return nil
// }

// func (test *Test) GetProjects() ([]models.Project, error) {

// 	var projects []models.Project
// 	projects = append([]models.Project{}, tests.NormalProjects...)

// 	for _, p := range projects {
// 		//removes IPs
// 		p.IPs = nil
// 	}

// 	return projects, nil
// }

// func (test *Test) GetProject(projectName string) (models.Project, error) {
// 	ps, err := test.GetProjects()
// 	var projects []models.Project
// 	projects = append([]models.Project{}, ps...)

// 	if err != nil {
// 		return models.Project{}, errors.New("Could not get projects" + err.Error())
// 	}

// 	for _, p := range projects {
// 		if p.Name == projectName {
// 			return p, nil
// 		}
// 	}
// 	return models.Project{}, errors.New("Could not get project " + projectName)
// }

// // IP
// func (test *Test) CreateOrUpdateIP(projectName string, ip models.IP) error {
// 	return nil
// }

// func (test *Test) CreateOrUpdateIPs(projectName string, ip []models.IP) error {
// 	return nil
// }

// func (test *Test) GetIPs(projectName string) ([]models.IP, error) {
// 	project, err := test.GetProject(projectName)

// 	var ips []models.IP
// 	if err != nil {
// 		return nil, errors.New("Could not get project : " + err.Error())
// 	}

// 	err = Clone(project.IPs, &ips)
// 	//remove ports !
// 	for i := range ips {
// 		if err != nil {
// 			return nil, errors.New("Could not clone IPs : " + err.Error())
// 		}
// 		ips = append(ips, project.IPs[i])
// 		ips[i].Ports = nil
// 	}

// 	fmt.Printf("SWDFHGJQSDFHGQSDJFGHQFDSJKGH ips %+v\n", ips)

// 	return ips, nil
// }

// func (test *Test) GetIP(projectName string, ip string) (models.IP, error) {

// 	ips, err := test.GetIPs(projectName)
// 	if err != nil {
// 		return models.IP{}, errors.New("Could not get IPs : " + err.Error())
// 	}

// 	var tmpIP models.IP
// 	var localIps []models.IP
// 	localIps = make([]models.IP, 0)

// 	for i := range ips {
// 		err = Clone(ips[i], &tmpIP)
// 		if err != nil {
// 			return models.IP{}, errors.New("Could not clone IPs : " + err.Error())
// 		}
// 		localIps = append(localIps, tmpIP)
// 		if localIps[i].Value == ip {
// 			// remove ports !
// 			localIps[i].Ports = nil
// 			return localIps[i], nil
// 		}
// 	}

// 	return models.IP{}, errors.New("IP not found")
// }

// // Domain
// func (test *Test) CreateOrUpdateDomain(projectName string, domain models.Domain) error {
// 	return errors.New("Not implemented yet")
// }

// func (test *Test) CreateOrUpdateDomains(projectName string, domain []models.Domain) error {
// 	return errors.New("Not implemented yet")
// }

// func (test *Test) GetDomains(projectName string) ([]models.Domain, error) {
// 	return tests.NormalProject.Domains, nil
// }

// func (test *Test) GetDomain(projectName string, domain string) (models.Domain, error) {
// 	return models.Domain{}, errors.New("Not implemented yet")
// }

// // Port
// func (test *Test) CreateOrUpdatePort(projectName string, ip string, port models.Port) error {
// 	return nil
// }

// func (test *Test) CreateOrUpdatePorts(projectName string, ip string, port []models.Port) error {
// 	return nil
// }

// func (test *Test) GetPorts(projectName string, ip string) ([]models.Port, error) {
// 	var ports []models.Port
// 	localIp, err := test.GetIP(projectName, ip)
// 	if err != nil {
// 		return nil, errors.New("Could not get ports : " + err.Error())
// 	}
// 	for _, p := range localIp.Ports {
// 		p.URIs = nil
// 		ports = append(ports, p)
// 	}
// 	//only one correponding ip !
// 	return ports, nil
// }

// func (test *Test) GetPort(projectName string, ip string, port string) (models.Port, error) {
// 	ports, err := test.GetPorts(projectName, ip)
// 	if err != nil {
// 		return models.Port{}, err
// 	}

// 	for _, p := range ports {
// 		if strconv.Itoa(int(p.Number)) == port {
// 			return p, nil
// 		}
// 	}
// 	return models.Port{}, errors.New("Port not found")
// }

// // URI (directory and files)
// func (test *Test) CreateOrUpdateURI(projectName string, ip string, port string, uri models.URI) error {
// 	return nil
// }

// func (test *Test) CreateOrUpdateURIs(projectName string, ip string, port string, uris []models.URI) error {
// 	return nil
// }

// func (test *Test) GetURIs(projectName string, ip string, port string) ([]models.URI, error) {
// 	p, err := test.GetPort(projectName, ip, port)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return p.URIs, nil
// }

// func (test *Test) GetURI(projectName string, ip string, port string, uri string) (models.URI, error) {
// 	uris, err := test.GetURIs(projectName, ip, port)
// 	if err != nil {
// 		return models.URI{}, nil
// 	}
// 	for _, u := range uris {
// 		if u.Name == uri {
// 			return u, nil
// 		}
// 	}
// 	return models.URI{}, errors.New("Uri not found")
// }

// // Raw data
// func (test *Test) AppendRawData(projectName string, moduleName string, data interface{}) error {
// 	return nil
// }

// func (test *Test) GetRaws(projectName string) (models.Raws, error) {
// 	raws, ok := tests.NormalRaws[projectName]
// 	if !ok {
// 		return models.Raws{}, errors.New("Project not found")
// 	}
// 	return raws, nil
// }

// func (test *Test) GetRawModule(projectName string, moduleName string) (map[string]interface{}, error) {
// 	raws, err := test.GetRaws(projectName)
// 	if err != nil {
// 		return nil, err
// 	}

// 	raw, ok := raws[moduleName]
// 	if !ok {
// 		return nil, errors.New("Module not found")
// 	}

// 	return raw, nil
// }
