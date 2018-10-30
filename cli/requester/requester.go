package requester

import (
	"github.com/mitchellh/mapstructure"
	"github.com/netm4ul/netm4ul/core/api"
	"github.com/netm4ul/netm4ul/core/database/models"
	"github.com/netm4ul/netm4ul/core/session"
	log "github.com/sirupsen/logrus"

	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

/*
getURL returns a full URL with all the configured attributes (ip, port, version...)
It takes a ressource name and a valid session
*/
func getURL(ressource string, s *session.Session) string {
	port := strconv.FormatInt(int64(s.Config.API.Port), 10)
	url := "http://" + s.Config.Server.IP + ":" + port + "/api/v1"

	if ressource != "" {
		url += ressource
	}

	return strings.TrimRight(url, "/")
}

/*
* getData takes care of getting data from the API
* It handles the creation of the full url, all the json formatting and verify the API return status
* Args:
	- the ressource name
	- the current sessions
* Return :
	- The returned data of the api using api.Result type
	- any error encountered during the execution
*/
func getData(ressource string, s *session.Session) (api.Result, error) {

	var result api.Result
	url := getURL(ressource, s)

	log.Debugf("GET : %s", url)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return api.Result{}, errors.New("Can't get create new request : " + err.Error())
	}

	req.Header.Set("X-Session-Token", s.Config.API.Token)

	res, err := client.Do(req)
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

/*
* postData takes care of posting data to the API
* It sets the correct content type and
* Args:
	- the ressource name
	- the current sessions
	- any kind of data that is expected by the api
* Return :
	- The returned data of the api using api.Result type
	- any error encountered during the execution
*/
func postData(ressource string, s *session.Session, rawdata interface{}) (api.Result, error) {
	var result api.Result

	jsondata, err := json.Marshal(rawdata)
	if err != nil {
		return api.Result{}, errors.New("Could not create project :" + err.Error())
	}
	url := getURL(ressource, s)

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsondata))
	if err != nil {
		return api.Result{}, errors.New("Can't get create new request : " + err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Session-Token", s.Config.API.Token)
	res, err := client.Do(req)

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
	p := models.Project{Name: s.Config.Project.Name, Description: s.Config.Project.Description}

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

/*
CreateProject is a wrapper function to create a new project.
*/
func CreateProject(p models.Project, s *session.Session) error {

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

/*
GetProjects is an helper function to get a slice of all the projects availables
Return :
	- Slice of models.Projects from any kind of database
	- error if anything unexpected occurred during the execution of the function
*/
func GetProjects(s *session.Session) ([]models.Project, error) {

	var data []models.Project
	resjson, err := getData("/projects", s)

	log.Debugf("response : %+v", resjson)

	if err != nil {
		return data, err
	}

	// using mapstructure to decode all the json response into the data variable.
	err = mapstructure.Decode(resjson.Data, &data)
	if err != nil {
		return data, err
	}

	// Check if the api response code say that everything went fine or abort.
	if resjson.Code != api.CodeOK {
		return data, errors.New("Can't get projects list :" + err.Error())
	}

	return data, nil
}

/*
GetProject is an helper function to get all the information from a project by its name
Return :
	- a models.Projects from any kind of database
	- error if anything unexpected occurred during the execution of the function
*/
func GetProject(name string, s *session.Session) (models.Project, error) {
	var data models.Project
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

func GetIPs(projectName string, s *session.Session) ([]models.IP, error) {
	var data []models.IP
	resjson, err := getData("/projects/"+projectName+"/ips", s)

	log.Debugf("response : %+v", resjson)
	if err != nil {
		return data, err
	}

	// using mapstructure to decode all the json response into the data variable.
	err = mapstructure.Decode(resjson.Data, &data)
	if err != nil {
		return data, err
	}

	// Check if the api response code say that everything went fine or abort.
	if resjson.Code != api.CodeOK {
		return data, errors.New("Can't get projects list :" + err.Error())
	}

	return data, nil
}

func GetDomains(projectName string, s *session.Session) ([]models.IP, error) {
	var data []models.IP
	resjson, err := getData("/projects/"+projectName+"/domains", s)

	log.Debugf("response : %+v", resjson)
	if err != nil {
		return data, err
	}

	// using mapstructure to decode all the json response into the data variable.
	err = mapstructure.Decode(resjson.Data, &data)
	if err != nil {
		return data, err
	}

	// Check if the api response code say that everything went fine or abort.
	if resjson.Code != api.CodeOK {
		return data, errors.New("Can't get projects list :" + err.Error())
	}

	return data, nil
}

func GetPorts(projectName string, ip string, s *session.Session) ([]models.Port, error) {
	var data []models.Port
	resjson, err := getData("/projects/"+projectName+"/ips/"+ip+"/ports", s)

	log.Debugf("response : %+v", resjson)
	if err != nil {
		return data, err
	}

	// using mapstructure to decode all the json response into the data variable.
	err = mapstructure.Decode(resjson.Data, &data)
	if err != nil {
		return data, err
	}

	// Check if the api response code say that everything went fine or abort.
	if resjson.Code != api.CodeOK {
		return data, errors.New("Can't get projects list :" + err.Error())
	}

	return data, nil
}

func GetURIs(projectName string, ip string, port string, s *session.Session) ([]models.URI, error) {
	var data []models.URI
	resjson, err := getData("/projects/"+projectName+"/ips/"+ip+"/ports/"+port+"/uris", s)

	log.Debugf("response : %+v", resjson)
	if err != nil {
		return data, err
	}

	// using mapstructure to decode all the json response into the data variable.
	err = mapstructure.Decode(resjson.Data, &data)
	if err != nil {
		return data, err
	}

	// Check if the api response code say that everything went fine or abort.
	if resjson.Code != api.CodeOK {
		return data, errors.New("Can't get projects list :" + err.Error())
	}

	return data, nil
}
