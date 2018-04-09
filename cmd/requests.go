package cli

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/netm4ul/netm4ul/cmd/colors"
	"github.com/netm4ul/netm4ul/core/api"
	"github.com/netm4ul/netm4ul/core/config"
	"github.com/netm4ul/netm4ul/core/server/database"
	"github.com/pkg/errors"
)

func getURL(ressource string) string {
	port := strconv.FormatInt(int64(config.Config.API.Port), 10)
	url := "http://" + config.Config.Server.IP + ":" + port

	if ressource != "" {
		url += ressource
	}

	return strings.TrimRight(url, "/")
}

func getData(ressource string) (api.Result, error) {

	var result api.Result
	url := getURL(ressource)

	if config.Config.Verbose {
		log.Println(colors.Yellow("GET : " + url))
	}

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

	if result.Code != 200 {
		return result, errors.New(result.Message)
	}

	return result, nil
}

func postData(ressource string, rawdata interface{}) (api.Result, error) {
	var result api.Result

	jsondata, err := json.Marshal(rawdata)
	if err != nil {
		return api.Result{}, errors.New("Could not create project :" + err.Error())
	}
	url := getURL(ressource)

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

func CreateProject(p database.Project) error {

	ressource := "/projects"

	res, err := postData(ressource, p)
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

func GetProjects() ([]database.Project, error) {

	var data []database.Project
	resjson, err := getData("/projects")

	if config.Config.Verbose {
		log.Printf(colors.Yellow("response : %+v"), resjson)
	}

	if err != nil {
		return data, err
	}

	err = mapstructure.Decode(resjson.Data, &data)
	if err != nil {
		return data, err
	}

	if resjson.Code != 200 {
		return data, errors.New("Can't get projects list :" + err.Error())
	}

	return data, nil
}

func GetProject(name string) (database.Project, error) {
	var data database.Project
	resjson, err := getData("/projects/" + name)

	if config.Config.Verbose {
		log.Printf(colors.Yellow("response : %+v"), resjson)
	}

	if err != nil {
		return data, err
	}

	err = mapstructure.Decode(resjson.Data, &data)
	if err != nil {
		return data, err
	}

	if resjson.Code != 200 {
		return data, errors.New("Can't get projects list :" + err.Error())
	}

	return data, nil
}

func GetIPsByProject(project string) (database.IP, error) {
	return database.IP{}, nil
}

func GetPortsByIP(project string, ip string) ([]database.Port, error) {
	return []database.Port{}, nil
}
