package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/netm4ul/netm4ul/modules"

	"github.com/netm4ul/netm4ul/core/session"
	log "github.com/sirupsen/logrus"

	"github.com/mitchellh/mapstructure"
	"github.com/netm4ul/netm4ul/core/api"
	"github.com/netm4ul/netm4ul/core/config"
	"github.com/netm4ul/netm4ul/core/database"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
)

func getURL(ressource string, s *session.Session) string {
	port := strconv.FormatInt(int64(s.Config.API.Port), 10)
	url := "http://" + s.Config.Server.IP + ":" + port

	if ressource != "" {
		url += ressource
	}

	return strings.TrimRight(url, "/")
}

func getData(ressource string, s *session.Session) (api.Result, error) {

	var result api.Result
	url := getURL(ressource, s)

	log.Debugf("GET : %s", url)

	res, err := http.Get(url)
	if err != nil {
		return api.Result{}, errors.New("Can't get " + ressource + " : " + err.Error())
	}
	defer res.Body.Close()

	// err = json.NewDecoder(res.Body).Decode(&result)
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return api.Result{}, errors.New("Can't read json : " + err.Error())
	}

	err = json.Unmarshal(data, &result)
	if err != nil {
		return api.Result{}, errors.New("Can't decode json : " + err.Error())
	}

	if result.Code != api.CodeOK {
		return result, errors.New(result.Message)
	}

	return result, nil
}

func postData(ressource string, s *session.Session, rawdata interface{}) (api.Result, error) {
	var result api.Result

	jsondata, err := json.Marshal(rawdata)
	if err != nil {
		return api.Result{}, errors.New("Could not create project :" + err.Error())
	}
	url := getURL(ressource, s)

	//TODO add header, auth
	res, err := http.Post(url, "application/json", bytes.NewBuffer(jsondata))
	if err != nil {
		return api.Result{}, errors.New("Could not post data " + err.Error())
	}

	err = json.NewDecoder(res.Body).Decode(result)
	if err != nil {
		return api.Result{}, errors.New("Could not read response data " + err.Error())
	}
	return result, nil
}

func createProjectIfNotExist(s *session.Session) {
	p := database.Project{Name: s.Config.Project.Name, Description: s.Config.Project.Description}

	listOfProject, err := GetProjects(s)
	if err != nil {
		log.Errorf("Can't get project list : %s", err.Error())
	}

	for _, project := range listOfProject {
		if project.Name == s.Config.Project.Name {
			return
			// already exist, so exit this function
		}
	}

	err = CreateProject(p, s)
	if err != nil {
		log.Errorf("Can't create project : %s", err.Error())
	}

}

func CreateProject(p database.Project, s *session.Session) error {

	ressource := "/projects"

	res, err := postData(ressource, s, p)
	if err != nil {
		return errors.New("Can't create project : %s")
	}

	if res.Data.(string) != p.Name {
		return errors.New("Error while creating the project")
	}
	return nil
}

type Projects struct {
	Projects []database.Project
}

func GetProjects(s *session.Session) ([]database.Project, error) {

	var data []database.Project
	resjson, err := getData("/projects", s)

	log.Debugf("response : %+v", resjson)

	if err != nil {
		return data, err
	}

	err = mapstructure.Decode(resjson.Data, &data)
	if err != nil {
		return data, err
	}

	if resjson.Code != api.CodeOK {
		return data, errors.New("Can't get projects list :" + err.Error())
	}

	return data, nil
}

func GetProject(name string, s *session.Session) (database.Project, error) {
	var data database.Project
	resjson, err := getData("/projects/"+name, s)

	log.Debugf("response : %+v", resjson)

	if err != nil {
		return data, err
	}

	err = mapstructure.Decode(resjson.Data, &data)
	if err != nil {
		return data, err
	}

	if resjson.Code != api.CodeOK {
		return data, errors.New("Can't get projects list :" + err.Error())
	}

	return data, nil
}

func GetIPsByProject(project string, s *session.Session) (database.IP, error) {
	return database.IP{}, nil
}

func GetPortsByIP(project string, ip string, s *session.Session) ([]database.Port, error) {
	return []database.Port{}, nil
}

func parseModules(modules []string, s *session.Session) ([]string, error) {

	if len(modules) == 0 {
		return nil, errors.New("Could not parse modules")
	}

	for _, name := range modules {
		_, ok := s.Config.Modules[name]
		if !ok {
			return nil, errors.New("Could not find module : " + name)
		}
	}

	return modules, nil
}

func addModules(mods []string, s *session.Session) {

	found := false
	for _, mod := range mods {
		for cmodname, cmod := range s.Config.Modules {
			if cmodname == mod && cmod.Enabled {
				found = true
			}
		}
		if !found {
			s.Config.Modules[mod] = config.Module{Enabled: true}
		}
	}
}

func printProjectsInfo(s *session.Session) {
	var err error
	var data [][]string

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Project", "Description", "# IPs", "Last update"})

	// get list of projects
	listOfProjects, err := GetProjects(s)
	if err != nil {
		log.Errorf("Can't get projects list : %s", err.Error())
	}

	// build array of array for the table !
	for _, p := range listOfProjects {
		if s.Config.Verbose {
			log.Infof("p : %+v", p)
		}
		data = append(data, []string{p.Name, p.Description, strconv.Itoa(len(p.IPs)), time.Unix(p.UpdatedAt, 0).String()})
	}

	table.AppendBulk(data)
	table.Render()
}

func printProjectInfo(projectName string, s *session.Session) {

	var p database.Project
	var err error
	var data [][]string

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"IP", "Ports"})

	if projectName == "" {
		log.Fatalln("No project provided")
		// exit
	}

	p, err = GetProject(projectName, s)
	if err != nil {
		log.Errorf("Can't get project %s : %s", projectName, err.Error())
	}

	log.Debugf("Project : %+v", p)

	for _, ip := range p.IPs {
		log.Debugf("ip : %+v", ip)
		for _, port := range ip.Ports {
			data = append(data, []string{ip.Value, strconv.Itoa(int(port.Number))})
		}
	}

	table.AppendBulk(data)
	table.Render()
}

func parseTargets(targets []string) ([]modules.Input, error) {

	var inputs []modules.Input
	var input modules.Input

	if len(targets) == 0 {
		return []modules.Input{}, errors.New("Not target found")
	}

	// loop on each targets
	for _, target := range targets {

		ip, ipNet, err := net.ParseCIDR(target)

		// if this is a domain
		if err != nil {
			ips, err := net.LookupIP(target)

			if err != nil {
				return []modules.Input{}, errors.New("Could lookup address : " + target + ", " + err.Error())
			}

			if ips == nil {
				return []modules.Input{}, errors.New("Could not resolve :" + target)
			}

			// convert ips to strings
			for _, ip := range ips {
				input = modules.Input{Domain: target, IP: ip}
				inputs = append(inputs, input)
			}

		} else {
			// if this is an ip
			// check if ip is specified (not :: or 0.0.0.0)
			if ip.IsUnspecified() {
				return []modules.Input{}, errors.New("Target ip is Unspecified (0.0.0.0 or ::)")
			}

			// check if ip isn't loopback
			if ip.IsLoopback() {
				return []modules.Input{}, errors.New("Target ip is loopback address")
			}

			// IP Range (CIDR)
			if ipNet != nil {
				h, err := hosts(target)
				if err != nil {
					return []modules.Input{}, errors.New("Target ip range is invalid (" + err.Error() + ")")
				}
				for _, host := range h {
					input = modules.Input{IP: host}
					inputs = append(inputs, input)
				}
			}
		}
	}

	return inputs, nil
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func hosts(cidr string) ([]net.IP, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []net.IP
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip)
	}
	// remove network address and broadcast address
	return ips[1 : len(ips)-1], nil
}

//PrintVersion Prints the version of all the components : The server, the Client, and the HTTP API
func PrintVersion(s *session.Session) {
	fmt.Printf("Version :\n - Server : %s\n - Client : %s\n - HTTP API : %s\n", s.Config.Versions.Server, s.Config.Versions.Client, s.Config.Versions.Api)
}
