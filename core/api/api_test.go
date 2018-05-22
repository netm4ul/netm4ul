package api_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/mitchellh/mapstructure"
	"github.com/netm4ul/netm4ul/core/database/models"
	"github.com/netm4ul/netm4ul/tests"

	"github.com/netm4ul/netm4ul/core/api"
	"github.com/netm4ul/netm4ul/core/config"
	"github.com/netm4ul/netm4ul/core/server"
	"github.com/netm4ul/netm4ul/core/session"
	log "github.com/sirupsen/logrus"
)

var (
	testserver *httptest.Server
	reader     io.Reader
	globalApi  *api.API
)

var conf config.ConfigToml

func init() {

	conf.Algorithm.Name = "random"
	conf.Project.Name = "Test project"
	conf.Project.Description = "Test description"
	conf.API.Port = 1234
	conf.Versions.Api = "v1"
	conf.Versions.Server = "vtestServer"
	conf.Versions.Client = "vtestClient"
	conf.Database.DatabaseType = "testadapter" // use local db, do not connect to external db (CI tests)

	sess, err := session.NewSession(conf)
	if err != nil {
		log.Fatalf("Could not create session %s", err)
	}
	serv := server.CreateServer(sess)

	globalApi = api.CreateAPI(sess, serv)
	globalApi.Routes()
}

func rrCheck(t *testing.T, method string, url string, handlerFunc http.HandlerFunc, httpCode int, apiCode api.Code, isLoggedIn bool) api.Result {
	var jsonres api.Result

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		t.Fatal(err)
	}

	//setup auth
	if isLoggedIn {
		req.Header.Add("X-Session-Token", tests.NormalUser.Token)
	}

	res := httptest.NewRecorder()

	globalApi.Router.ServeHTTP(res, req)

	err = json.NewDecoder(res.Body).Decode(&jsonres)
	if err != nil {
		t.Errorf("Could not decode json ! : %+v", err)
	}

	if res.Code != httpCode {
		t.Errorf("Bad HTTP Code, got %d instead of %d", res.Code, httpCode)
	}

	if jsonres.Code != apiCode {
		t.Errorf("Bad response code, got %d instead of %d", jsonres.Code, apiCode)
	}

	return jsonres
}

func TestHTTPIndex(t *testing.T) {
	url := globalApi.Prefix + "/"
	jsonres := rrCheck(t, "GET", url, globalApi.GetIndex, http.StatusOK, api.CodeOK, false)

	expected := "Documentation available at https://github.com/netm4ul/netm4ul"
	if jsonres.Message != expected {
		t.Errorf("Expected %s, got %s", expected, jsonres.Message)
	}
}

func checkAuth(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
	t, err := route.GetPathTemplate()
	methods, err := route.GetMethods()
	if err != nil {
		return err
	}
	url := strings.Replace(t, "{name}", "test", -1)
	log.Debugf("Test auth against : %s %s", methods[0], url)

	//TOFIX
	// ignore non GET method for now
	if methods[0] != "GET" {
		return nil
	}
	req, err := http.NewRequest(methods[0], url, nil)

	if err != nil {
		return err
	}

	res := httptest.NewRecorder()

	globalApi.Router.ServeHTTP(res, req)

	return nil
}

func TestAPI_Auth(t *testing.T) {
	globalApi.Router.Walk(checkAuth)
}

func TestAPI_GetProjects(t *testing.T) {
	url := globalApi.Prefix + "/projects"
	jsonres := rrCheck(t, "GET", url, globalApi.GetProjects, http.StatusOK, api.CodeOK, true)

	//Do not continue if failed !
	if jsonres.Data == nil {
		t.Fatal("Got empty data")
	}
	var projects []models.Project
	mapstructure.Decode(jsonres.Data, &projects)

	if projects[0].Name != conf.Project.Name && projects[0].Description == conf.Project.Description {
		t.Errorf("Expected name : %s, got %s", conf.Project.Name, projects[0].Name)
	}

	if projects[0].Description != conf.Project.Description {
		t.Errorf("Expected description : %s, got %s", conf.Project.Description, projects[0].Description)
	}
}
func TestAPI_GetProject(t *testing.T) {
	url := globalApi.Prefix + "/projects/" + tests.NormalProject.Name
	jsonres := rrCheck(t, "GET", url, globalApi.GetProject, http.StatusOK, api.CodeOK, true)

	//Do not continue if failed !
	if jsonres.Data == nil {
		t.Fatal("Got empty data")
	}
	var project models.Project
	mapstructure.Decode(jsonres.Data, &project)

	if project.Name != tests.NormalProject.Name {
		t.Errorf("Expected name : %s, got %s", tests.NormalProject.Name, project.Name)
	}

	if project.Description != tests.NormalProject.Description {
		t.Errorf("Expected description : %s, got %s", tests.NormalProject.Description, project.Description)
	}
}

func TestAPI_GetUser(t *testing.T) {
	url := globalApi.Prefix + "/users/" + tests.NormalUser.Name
	jsonres := rrCheck(t, "GET", url, globalApi.GetUser, http.StatusOK, api.CodeOK, true)

	//Do not continue if failed !
	if jsonres.Data == nil {
		t.Fatal("Got empty data")
	}

	var user models.User
	mapstructure.Decode(jsonres.Data, &user)

	if user.Name != tests.NormalUser.Name {
		t.Errorf("Expected description : %s, got %s", tests.NormalUser.Name, user.Name)
	}

	// check for sensitive information !
	if user.ID != "" {
		t.Errorf("User ID available !")
	}

	if user.Password != "" {
		t.Errorf("Password available !")
	}

	if user.Token != "" {
		t.Errorf("Token available !")
	}
}

func TestAPI_GetAlgorithm(t *testing.T) {
	url := globalApi.Prefix + "/projects/" + conf.Project.Name + "/algorithm"
	jsonres := rrCheck(t, "GET", url, globalApi.GetProjects, http.StatusOK, api.CodeOK, true)

	//Do not continue if failed !
	if jsonres.Data == nil {
		t.Fatal("Got empty data")
	}

	if jsonres.Data == conf.Algorithm {
		t.Errorf("Expected %s, got %s", conf.Algorithm, jsonres.Data)
	}
}
