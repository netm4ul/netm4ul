package api_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/netm4ul/netm4ul/core/database/models"

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

	conf.Versions.Api = "vtestApi"
	conf.Versions.Server = "vtestServer"
	conf.Versions.Client = "vtestClient"
	conf.Database.DatabaseType = "testadapter" // use local db, do not connect to external db (CI tests)

	sess, err := session.NewSession(conf)
	if err != nil {
		log.Fatalf("Could not create session %s", err)
	}
	serv := server.CreateServer(sess)

	globalApi = api.CreateAPI(sess, serv)

}

func rrCheck(t *testing.T, url string, handlerFunc http.HandlerFunc, httpCode int, apiCode api.Code) api.Result {
	var jsonres api.Result

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlerFunc)

	handler.ServeHTTP(rr, req)

	err = json.NewDecoder(rr.Body).Decode(&jsonres)

	if err != nil {
		t.Errorf("Could not decode json ! : %+v", err)
	}

	if rr.Code != http.StatusOK {
		t.Errorf("Bad HTTP Code, got %d instead of %d", rr.Code, http.StatusOK)
	}

	if jsonres.Code != api.CodeOK {
		t.Errorf("Bad response code, got %d instead of %d", jsonres.Code, api.CodeOK)
	}

	return jsonres
}

func TestHTTPIndex(t *testing.T) {
	url := "/api/" + conf.Versions.Api
	jsonres := rrCheck(t, url, globalApi.GetIndex, http.StatusOK, api.CodeOK)

	expected := "Documentation available at https://github.com/netm4ul/netm4ul"
	if jsonres.Message != expected {
		t.Errorf("Expected %s, got %s", expected, jsonres.Message)
	}
}

func TestAPI_GetProjects(t *testing.T) {
	url := "/api/" + conf.Versions.Api + "/projects"
	jsonres := rrCheck(t, url, globalApi.GetProjects, http.StatusOK, api.CodeOK)

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

func TestAPI_GetAlgorithm(t *testing.T) {
	url := "/api/" + conf.Versions.Api + "/projects/" + conf.Project.Name + "/algorithm"
	jsonres := rrCheck(t, url, globalApi.GetProjects, http.StatusOK, api.CodeOK)

	//Do not continue if failed !
	if jsonres.Data == nil {
		t.Fatal("Got empty data")
	}

	if jsonres.Data == conf.Algorithm {
		t.Errorf("Expected %s, got %s", conf.Algorithm, jsonres.Data)
	}
}
