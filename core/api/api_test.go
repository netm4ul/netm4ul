package api_test

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/netm4ul/netm4ul/core/api"
	"github.com/netm4ul/netm4ul/core/config"
	"github.com/netm4ul/netm4ul/core/database/models"
	"github.com/netm4ul/netm4ul/core/server"
	"github.com/netm4ul/netm4ul/core/session"
	"github.com/netm4ul/netm4ul/tests"
	log "github.com/sirupsen/logrus"
)

var (
	testserver *httptest.Server
	reader     io.Reader
)

func setup(conf config.ConfigToml) *api.API {
	sess, err := session.NewSession(conf)
	if err != nil {
		log.Fatalf("Could not create session %s", err)
	}
	serv := server.CreateServer(sess)

	a := api.NewAPI(sess, serv)
	a.Routes()
	return a
}

//rrCheck : checks the response code for the provided request
func rrCheck(t *testing.T, localApi *api.API, method string, url string, body io.Reader, handlerFunc http.HandlerFunc, httpCode int, apiCode api.Code, isLoggedIn bool) api.Result {
	var jsonres api.Result
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		t.Fatal(err)
	}

	//setup auth
	if isLoggedIn {
		req.Header.Add("X-Session-Token", tests.NormalUser.Token)
	}

	res := httptest.NewRecorder()

	localApi.Router.ServeHTTP(res, req)
	t.Logf("res.Body : %+v", res.Body)
	err = json.NewDecoder(res.Body).Decode(&jsonres)
	if err != nil {
		t.Errorf("Could not decode json ! : %+v", err)
	}

	if res.Code != httpCode {
		t.Errorf("Bad HTTP Code, got %d instead of %d", res.Code, httpCode)
	}

	if jsonres.Code != apiCode {
		t.Errorf("Bad response code, got \"%s\" instead of \"%s\"", jsonres.Code, apiCode)
	}
	return jsonres
}

func TestNewAPI(t *testing.T) {
	type args struct {
		s      *session.Session
		server *server.Server
	}
	fakeSessionIPv4 := session.Session{
		Config: config.ConfigToml{
			API: config.API{
				IP:   net.IPv4(0, 0, 0, 0).String(),
				Port: 8888,
			},
		},
	}
	fakeSessionIPv6 := session.Session{
		Config: config.ConfigToml{
			API: config.API{
				IP:   net.IPv6interfacelocalallnodes.String(),
				Port: 8888,
			},
		},
	}
	fakeServer := server.Server{}

	t.Run("Test API constructor initialization IPv4", func(t *testing.T) {
		a := api.NewAPI(&fakeSessionIPv4, &fakeServer)
		if a.IPPort != fakeSessionIPv4.Config.API.IP+":"+strconv.Itoa(int(fakeSessionIPv4.Config.API.Port)) {
			t.Errorf("IPPort mismatch : want %s, got %s", fakeSessionIPv4.Config.API.IP+":"+strconv.Itoa(int(fakeSessionIPv4.Config.API.Port)), a.IPPort)
		}
	})

	t.Run("Test API constructor initialization IPv6", func(t *testing.T) {
		a := api.NewAPI(&fakeSessionIPv6, &fakeServer)
		if a.IPPort != "[ff01::1]:8888" {
			t.Errorf("IPPort mismatch : want %s, got %s", "[ff01::1]:8888", a.IPPort)
		}
	})
}
func TestHTTPIndex(t *testing.T) {
	conf := config.ConfigToml{
		API: config.API{
			Port: 1234,
			IP:   "0.0.0.0",
		},
		Database: config.Database{
			DatabaseType: "testadapter",
		},
		Algorithm: config.Algorithm{
			Name: "random",
		},
	}

	localApi := setup(conf)
	url := localApi.Prefix + "/"
	jsonres := rrCheck(t, localApi, "GET", url, nil, localApi.GetIndex, http.StatusOK, api.CodeOK, false)

	expected := "Documentation available at https://github.com/netm4ul/netm4ul"
	if jsonres.Message != expected {
		t.Errorf("Expected %s, got %s", expected, jsonres.Message)
	}
}

func TestAPI_Auth(t *testing.T) {
	// TODO

	// conf := config.ConfigToml{
	// 	API: config.API{
	// 		Port: 1234,
	// 		IP: "0.0.0.0",
	// 	},
	// 	Database: config.Database{
	// 		DatabaseType: "testadapter",
	// 	},
	// }

	// localApi := setup(conf)
}

func TestAPI_GetProjects(t *testing.T) {
	conf := config.ConfigToml{
		API: config.API{
			Port: 1234,
			IP:   "0.0.0.0",
		},
		Database: config.Database{
			DatabaseType: "testadapter",
		},
		Algorithm: config.Algorithm{
			Name: "random",
		},
	}
	var localApi *api.API
	var url string
	var jsonres api.Result
	var projects []models.Project

	localApi = setup(conf)

	url = localApi.Prefix + "/projects"
	jsonres = rrCheck(t, localApi, "GET", url, nil, localApi.GetProjects, http.StatusOK, api.CodeOK, true)

	//Do not continue if failed !
	if jsonres.Data == nil {
		t.Error("Got empty data")
	}

	err := mapstructure.Decode(jsonres.Data, &projects)
	if err != nil {
		t.Errorf("Could not decode JSON : %s", err)
	}

	for i, project := range projects {
		if project.Name != tests.NormalProjects[i].Name {
			t.Errorf("Expected name : %s, got %s", tests.NormalProjects, project.Name)
		}

		if project.Description != tests.NormalProjects[i].Description {
			t.Errorf("Expected description : %s, got %s", tests.NormalProjects, project.Description)
		}
	}

}
func TestAPI_GetProject(t *testing.T) {
	conf := config.ConfigToml{
		API: config.API{
			Port: 1234,
			IP:   "0.0.0.0",
		},
		Database: config.Database{
			DatabaseType: "testadapter",
		},
		Algorithm: config.Algorithm{
			Name: "random",
		},
	}

	localApi := setup(conf)
	url := localApi.Prefix + "/projects/" + tests.NormalProject.Name
	jsonres := rrCheck(t, localApi, "GET", url, nil, localApi.GetProject, http.StatusOK, api.CodeOK, true)

	//Do not continue if failed !
	if jsonres.Data == nil {
		t.Error("Got empty data")
	}
	var project models.Project
	err := mapstructure.Decode(jsonres.Data, &project)
	if err != nil {
		t.Fatalf("Could not decode JSON : %s", err)
	}
	if project.Name != tests.NormalProject.Name {
		t.Errorf("Expected name : %s, got %s", tests.NormalProject.Name, project.Name)
	}

	if project.Description != tests.NormalProject.Description {
		t.Errorf("Expected description : %s, got %s", tests.NormalProject.Description, project.Description)
	}
}

func TestAPI_GetUser(t *testing.T) {
	conf := config.ConfigToml{
		API: config.API{
			Port: 1234,
			IP:   "0.0.0.0",
		},
		Database: config.Database{
			DatabaseType: "testadapter",
		},
		Algorithm: config.Algorithm{
			Name: "random",
		},
	}

	localApi := setup(conf)

	t.Run("Check the getting existing user informations", func(t *testing.T) {
		url := localApi.Prefix + "/users/" + tests.NormalUser.Name
		jsonres := rrCheck(t, localApi, "GET", url, nil, localApi.GetUser, http.StatusOK, api.CodeOK, true)

		//Do not continue if failed !
		if jsonres.Data == nil {
			t.Error("Got empty data")
		}

		var user models.User
		err := mapstructure.Decode(jsonres.Data, &user)
		if err != nil {
			t.Errorf("Could not decode JSON : %s", err)
		}

		if user.Name != tests.NormalUser.Name {
			t.Errorf("Expected description : %s, got %s", tests.NormalUser.Name, user.Name)
		}

		if user.Password != "" {
			t.Errorf("Password available !")
		}

		if user.Token != "" {
			t.Errorf("Token available !")
		}
	})

	t.Run("Check the getting non-existing user informations", func(t *testing.T) {
		url := localApi.Prefix + "/users/" + "nonExistingUser"
		jsonres := rrCheck(t, localApi, "GET", url, nil, localApi.GetUser, http.StatusNotFound, api.CodeNotFound, true)

		if jsonres.Data != nil {
			t.Error("The api returned some data and it shouldn't have")
		}
	})
}

func TestAPI_GetAlgorithm(t *testing.T) {
	conf := config.ConfigToml{
		API: config.API{
			Port: 1234,
			IP:   "0.0.0.0",
		},
		Project: config.Project{
			Name: "test",
		},
		Database: config.Database{
			DatabaseType: "testadapter",
		},
		Algorithm: config.Algorithm{
			Name: "random",
		},
	}

	localApi := setup(conf)

	t.Run("Check the getting the algorithm", func(t *testing.T) {
		url := localApi.Prefix + "/projects/" + conf.Project.Name + "/algorithm"
		jsonres := rrCheck(t, localApi, "GET", url, nil, localApi.GetProjects, http.StatusOK, api.CodeOK, true)

		//Do not continue if failed !
		if jsonres.Data == nil {
			t.Error("Got empty data")
		}

		if jsonres.Data == conf.Algorithm {
			t.Errorf("Expected %s, got %s", conf.Algorithm, jsonres.Data)
		}
	})

}
func TestAPI_GetIPsByProjectName(t *testing.T) {
	conf := config.ConfigToml{
		API: config.API{
			Port: 1234,
			IP:   "0.0.0.0",
		},
		Project: config.Project{
			Name: "test",
		},
		Database: config.Database{
			DatabaseType: "testadapter",
		},
		Algorithm: config.Algorithm{
			Name: "random",
		},
	}

	localApi := setup(conf)

	t.Run("Check the getting IP for an empty project (no ip)", func(t *testing.T) {
		backup := tests.NormalIPs
		tests.NormalIPs = nil
		url := localApi.Prefix + "/projects/" + conf.Project.Name + "/ips"
		jsonres := rrCheck(t, localApi, "GET", url, nil, localApi.GetIPsByProjectName, http.StatusNotFound, api.CodeNotFound, true)
		if jsonres.Data != nil {
			t.Errorf("Got data (%s), should be nil", jsonres.Data)
		}
		tests.NormalIPs = backup
	})

	t.Run("Check the getting IP for a project", func(t *testing.T) {
		var ips []models.IP
		url := localApi.Prefix + "/projects/" + conf.Project.Name + "/ips"
		jsonres := rrCheck(t, localApi, "GET", url, nil, localApi.GetIPsByProjectName, http.StatusOK, api.CodeOK, true)

		err := mapstructure.Decode(jsonres.Data, &ips)
		if err != nil {
			t.Errorf("Could not decode JSON : %s", err)
		}
		if len(ips) == 0 {
			t.Error("No ip found")
		}

		for i, ip := range ips {
			if ip.Value != tests.NormalIPs[i].Value {
				t.Errorf("Received wrong value for ,got %s instead of %s", ip.Value, tests.NormalIPs[i].Value)
			}
			if ip.Network != tests.NormalIPs[i].Network {
				t.Errorf("Received wrong value for ,got %s instead of %s", ip.Network, tests.NormalIPs[i].Network)
			}
			if ip.CreatedAt != tests.NormalIPs[i].CreatedAt {
				t.Errorf("Received wrong value for ,got %s instead of %s", ip.CreatedAt, tests.NormalIPs[i].CreatedAt)
			}
			if ip.UpdatedAt != tests.NormalIPs[i].UpdatedAt {
				t.Errorf("Received wrong value for ,got %s instead of %s", ip.UpdatedAt, tests.NormalIPs[i].UpdatedAt)
			}
		}
	})
}

func TestAPI_ChangeAlgorithm(t *testing.T) {
	conf := config.ConfigToml{
		API: config.API{
			Port: 1234,
			IP:   "0.0.0.0",
		},
		Project: config.Project{
			Name: "test",
		},
		Database: config.Database{
			DatabaseType: "testadapter",
		},
		Algorithm: config.Algorithm{
			Name: "random",
		},
	}
	localApi := setup(conf)

	t.Run("Check the changing algorithm function", func(t *testing.T) {
		urlChangeAlgo := localApi.Prefix + "/projects/" + conf.Project.Name + "/algorithm"
		body := strings.NewReader("\"roundrobin\"")
		jsonres := rrCheck(t, localApi, "POST", urlChangeAlgo, body, localApi.ChangeAlgorithm, http.StatusOK, api.CodeOK, true)

		expected := "Algorithm changed to : roundrobin"
		if jsonres.Message != expected {
			t.Errorf("Got the wrong response message : [%s] instead of [%s]", jsonres.Message, expected)
		}

		if jsonres.Data == conf.Algorithm {
			t.Errorf("Expected %s, got %s", conf.Algorithm, jsonres.Data)
		}
	})

	t.Run("Check if the changes propagate correctly", func(t *testing.T) {
		urlGetAlgo := localApi.Prefix + "/projects/" + conf.Project.Name + "/algorithm"
		jsonresAfterChange := rrCheck(t, localApi, "GET", urlGetAlgo, nil, localApi.GetAlgorithm, http.StatusOK, api.CodeOK, true)

		expected := "roundrobin"
		if strings.ToLower(jsonresAfterChange.Data.(string)) != expected {
			t.Errorf("Got the wrong response data : [%s] instead of [%s]", jsonresAfterChange.Data, expected)
		}

		if jsonresAfterChange.Data == conf.Algorithm {
			t.Errorf("Expected %s, got %s", conf.Algorithm, jsonresAfterChange.Data)
		}
	})
	t.Run("Check the changing algorithm to an unknown one", func(t *testing.T) {
		urlChangeAlgo := localApi.Prefix + "/projects/" + conf.Project.Name + "/algorithm"
		body := strings.NewReader("\"NON-EXISTING-ALGORITHM\"")
		jsonres := rrCheck(t, localApi, "POST", urlChangeAlgo, body, localApi.ChangeAlgorithm, http.StatusUnprocessableEntity, api.CodeInvalidInput, true)

		if jsonres.Data != nil {
			t.Errorf("Expected empty data, got %s", jsonres.Data)
		}

		urlGetAlgo := localApi.Prefix + "/projects/" + conf.Project.Name + "/algorithm"
		jsonresAfterChange := rrCheck(t, localApi, "GET", urlGetAlgo, nil, localApi.GetAlgorithm, http.StatusOK, api.CodeOK, true)
		expected := "roundrobin"
		if strings.ToLower(jsonresAfterChange.Data.(string)) != expected {
			t.Errorf("Got the wrong algorithm : got [%s] instead of [%s]", jsonresAfterChange.Data, expected)
		}
	})

	t.Run("Sending invalid json data", func(t *testing.T) {
		urlChangeAlgo := localApi.Prefix + "/projects/" + conf.Project.Name + "/algorithm"
		body := strings.NewReader("INVALID-JSON")
		jsonres := rrCheck(t, localApi, "POST", urlChangeAlgo, body, localApi.ChangeAlgorithm, http.StatusBadRequest, api.CodeCouldNotDecodeJSON, true)

		if jsonres.Data != nil {
			t.Errorf("Expected empty data, got %s", jsonres.Data)
		}

		urlGetAlgo := localApi.Prefix + "/projects/" + conf.Project.Name + "/algorithm"
		jsonresAfterChange := rrCheck(t, localApi, "GET", urlGetAlgo, nil, localApi.GetAlgorithm, http.StatusOK, api.CodeOK, true)
		expected := "roundrobin"
		if strings.ToLower(jsonresAfterChange.Data.(string)) != expected {
			t.Errorf("Got the wrong algorithm : got [%s] instead of [%s]", jsonresAfterChange.Data, expected)
		}
	})
}

func TestAPI_GetPortsByIP(t *testing.T) {
	conf := config.ConfigToml{
		API: config.API{
			Port: 1234,
			IP:   "0.0.0.0",
		},
		Project: config.Project{
			Name: "test",
		},
		Database: config.Database{
			DatabaseType: "testadapter",
		},
		Algorithm: config.Algorithm{
			Name: "random",
		},
	}

	localApi := setup(conf)
	t.Run("Get ports from an empty ip (no ports)", func(t *testing.T) {
		backup := tests.NormalPorts
		tests.NormalPorts = []models.Port{}
		url := localApi.Prefix + "/projects/" + conf.Project.Name + "/ips/" + tests.NormalIPs[0].Value + "/ports"
		jsonres := rrCheck(t, localApi, "GET", url, nil, localApi.GetPortsByIP, http.StatusNotFound, api.CodeNotFound, true)
		if jsonres.Data != nil {
			t.Errorf("Got data (%s), should be nil", jsonres.Data)
		}
		tests.NormalPorts = backup
	})

	t.Run("Get all the ports for an IP", func(t *testing.T) {
		var ports []models.Port
		url := localApi.Prefix + "/projects/" + conf.Project.Name + "/ips/" + tests.NormalIPs[0].Value + "/ports"
		t.Logf("URL : %s", url)
		jsonres := rrCheck(t, localApi, "GET", url, nil, localApi.GetPortsByIP, http.StatusOK, api.CodeOK, true)

		err := mapstructure.Decode(jsonres.Data, &ports)
		if err != nil {
			t.Errorf("Could not decode JSON : %s", err)
		}

		if len(ports) == 0 {
			t.Error("No ports found")
		}

		for i, port := range ports {
			if port.Type != tests.NormalPorts[i].Type {
				t.Errorf("Received wrong Type for ,got %s instead of %s", port.Type, tests.NormalPorts[i].Type)
			}
			if port.Status != tests.NormalPorts[i].Status {
				t.Errorf("Received wrong value for ,got %s instead of %s", port.Status, tests.NormalPorts[i].Status)
			}
			if port.Protocol != tests.NormalPorts[i].Protocol {
				t.Errorf("Received wrong value for ,got %s instead of %s", port.Protocol, tests.NormalPorts[i].Protocol)
			}
			if port.Banner != tests.NormalPorts[i].Banner {
				t.Errorf("Received wrong value for ,got %s instead of %s", port.Banner, tests.NormalPorts[i].Banner)
			}
			if port.Number != tests.NormalPorts[i].Number {
				t.Errorf("Received wrong value for ,got %d instead of %d", port.Number, tests.NormalPorts[i].Number)
			}
			if port.CreatedAt != tests.NormalPorts[i].CreatedAt {
				t.Errorf("Received wrong value for ,got %s instead of %s", port.CreatedAt, tests.NormalPorts[i].CreatedAt)
			}
			if port.UpdatedAt != tests.NormalPorts[i].UpdatedAt {
				t.Errorf("Received wrong value for ,got %s instead of %s", port.UpdatedAt, tests.NormalIPs[i].UpdatedAt)
			}
		}
	})

}

func TestAPI_GetURIByPort(t *testing.T) {
	conf := config.ConfigToml{
		API: config.API{
			Port: 1234,
			IP:   "0.0.0.0",
		},
		Project: config.Project{
			Name: "test",
		},
		Database: config.Database{
			DatabaseType: "testadapter",
		},
		Algorithm: config.Algorithm{
			Name: "random",
		},
	}

	localApi := setup(conf)

	t.Run("Get URI from a new port (empty)", func(t *testing.T) {
		backup := tests.NormalURIs
		tests.NormalURIs = []models.URI{}
		urlGetUri := localApi.Prefix + "/projects/" + conf.Project.Name + "/ips/" + tests.NormalIPs[0].Value + "/ports/" + strconv.Itoa(int(tests.NormalPorts[0].Number)) + "/tcp/uris"
		jsonres := rrCheck(t, localApi, "GET", urlGetUri, nil, localApi.GetURIByPort, http.StatusOK, api.CodeOK, true)
		if jsonres.Data != nil {
			t.Errorf("Got data (%s), should be nil", jsonres.Data)
		}
		tests.NormalURIs = backup
	})
}

func TestAPI_GetRawModuleByProject(t *testing.T) {
	conf := config.ConfigToml{
		API: config.API{
			Port: 1234,
			IP:   "0.0.0.0",
		},
		Project: config.Project{
			Name: "test",
		},
		Database: config.Database{
			DatabaseType: "testadapter",
		},
		Algorithm: config.Algorithm{
			Name: "random",
		},
	}

	_ = setup(conf)
	t.Run("Get Raws from a new project (empty)", func(t *testing.T) {

	})
}

func TestAPI_GetRoutesByIP(t *testing.T) {
	conf := config.ConfigToml{
		API: config.API{
			Port: 1234,
			IP:   "0.0.0.0",
		},
		Project: config.Project{
			Name: "test",
		},
		Database: config.Database{
			DatabaseType: "testadapter",
		},
		Algorithm: config.Algorithm{
			Name: "random",
		},
	}

	_ = setup(conf)
	t.Run("Get ports from an empty ip (no ports)", func(t *testing.T) {
		
	})
}

func TestAPI_CreateProject(t *testing.T) {
	conf := config.ConfigToml{
		API: config.API{
			Port: 1234,
			IP:   "0.0.0.0",
		},
		Project: config.Project{
			Name: "test",
		},
		Database: config.Database{
			DatabaseType: "testadapter",
		},
		Algorithm: config.Algorithm{
			Name: "random",
		},
	}

	_ = setup(conf)
	t.Run("Get ports from an empty ip (no ports)", func(t *testing.T) {
		
	})
}

func TestAPI_RunModules(t *testing.T) {
	conf := config.ConfigToml{
		API: config.API{
			Port: 1234,
			IP:   "0.0.0.0",
		},
		Project: config.Project{
			Name: "test",
		},
		Database: config.Database{
			DatabaseType: "testadapter",
		},
		Algorithm: config.Algorithm{
			Name: "random",
		},
	}

	_ = setup(conf)
	t.Run("Get ports from an empty ip (no ports)", func(t *testing.T) {
		
	})
}

func TestAPI_RunModule(t *testing.T) {
	conf := config.ConfigToml{
		API: config.API{
			Port: 1234,
			IP:   "0.0.0.0",
		},
		Project: config.Project{
			Name: "test",
		},
		Database: config.Database{
			DatabaseType: "testadapter",
		},
		Algorithm: config.Algorithm{
			Name: "random",
		},
	}

	_ = setup(conf)
	t.Run("Get ports from an empty ip (no ports)", func(t *testing.T) {
		
	})
}

func TestAPI_DeleteProject(t *testing.T) {
	conf := config.ConfigToml{
		API: config.API{
			Port: 1234,
			IP:   "0.0.0.0",
		},
		Project: config.Project{
			Name: "test",
		},
		Database: config.Database{
			DatabaseType: "testadapter",
		},
		Algorithm: config.Algorithm{
			Name: "random",
		},
	}

	_ = setup(conf)
	t.Run("Get ports from an empty ip (no ports)", func(t *testing.T) {
		
	})
}
