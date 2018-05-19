package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/netm4ul/netm4ul/modules"

	"github.com/gorilla/mux"
	"github.com/netm4ul/netm4ul/core/communication"
	"github.com/netm4ul/netm4ul/core/config"
	"github.com/netm4ul/netm4ul/core/database/models"
	"github.com/netm4ul/netm4ul/core/server"
	"github.com/netm4ul/netm4ul/core/session"
	log "github.com/sirupsen/logrus"
)

// Result is the standard response format
type Result struct {
	Status   string      `json:"status"`
	Code     Code        `json:"code"`
	Message  string      `json:"message,omitempty"`
	Data     interface{} `json:"data,omitempty"`
	HTTPCode int         `json:"-"` //remove HTTPCode from the json response
}

//API is the constructor for this package
type API struct {
	// Session defines the global session for the API.
	Session *session.Session
	Server  *server.Server
	db      models.Database
}

//Info provides general purpose information for this API
type Info struct {
	Port     uint16          `json:"port,omitempty"`
	Versions config.Versions `json:"versions"`
}

//Metadata of the current system (node, api, database)
type Metadata struct {
	Nodes []communication.Node `json:"nodes"`
	Info  Info                 `json:"api"`
}

// CreateAPI : Initialise the infinite server loop on the master node
func CreateAPI(s *session.Session, server *server.Server) *API {
	api := API{
		Session: s,
		Server:  server,
		db:      server.Db,
	}

	return &api
}

//Start the API and route endpoints to functions
func (api *API) Start() {

	ipport := api.Session.GetAPIIPPort()
	router := api.Handler()
	log.Fatal(http.ListenAndServe(ipport, router))
}

//Handler return a new mux router. All
func (api *API) Handler() *mux.Router {

	ipport := api.Session.GetAPIIPPort()
	version := api.Session.Config.Versions.Api
	prefix := "/api/" + version

	log.Infof("API Listenning : %s, version : %s", ipport, version)
	log.Infof("API Endpoint : %s", ipport+prefix)

	router := mux.NewRouter()

	// Add content-type json header !
	router.Use(jsonMiddleware)

	// GET
	router.HandleFunc(prefix+"/", api.GetIndex).Methods("GET")
	router.HandleFunc(prefix+"/projects", api.GetProjects).Methods("GET")
	router.HandleFunc(prefix+"/projects/{name}", api.GetProject).Methods("GET")
	router.HandleFunc(prefix+"/projects/{name}/algorithm", api.GetAlgorithm).Methods("GET")
	router.HandleFunc(prefix+"/projects/{name}/ips", api.GetIPsByProjectName).Methods("GET")
	router.HandleFunc(prefix+"/projects/{name}/ips/{ip}/ports", api.GetPortsByIP).Methods("GET")            // We don't need to go deeper. Get all ports at once
	router.HandleFunc(prefix+"/projects/{name}/ips/{ip}/ports/{protocol}", api.GetPortsByIP).Methods("GET") // get only one protocol result (tcp, udp). Same GetPortsByIP function
	router.HandleFunc(prefix+"/projects/{name}/ips/{ip}/ports/{protocol}/{port}/directories", api.GetURIByPort).Methods("GET")
	router.HandleFunc(prefix+"/projects/{name}/ips/{ip}/routes", api.GetRoutesByIP).Methods("GET")
	router.HandleFunc(prefix+"/projects/{name}/raw/{module}", api.GetRawModuleByProject).Methods("GET")

	// POST
	router.HandleFunc(prefix+"/projects", api.CreateProject).Methods("POST")
	router.HandleFunc(prefix+"/projects/{name}/algorithm", api.ChangeAlgorithm).Methods("POST")
	router.HandleFunc(prefix+"/projects/{name}/run", api.RunModules).Methods("POST")
	router.HandleFunc(prefix+"/projects/{name}/run/{module}", api.RunModule).Methods("POST")

	// DELETE
	router.HandleFunc(prefix+"/projects/{name}", api.DeleteProject).Methods("DELETE")
	return router
}

//GetIndex returns a link to the documentation on the root path
func (api *API) GetIndex(w http.ResponseWriter, r *http.Request) {

	info := Info{Port: api.Session.Config.API.Port, Versions: api.Session.Config.Versions}
	d := Metadata{Info: info, Nodes: api.Server.Session.Nodes}

	res := CodeToResult[CodeOK]
	res.Data = d
	res.Message = "Documentation available at https://github.com/netm4ul/netm4ul"
	json.NewEncoder(w).Encode(res)
}

/*
GetProjects return this template
  "data": [
    {
	  "name": "FirstProject",
	  "description": "Some description",
	  "updated_at": 12345678
    }
  ]
*/
func (api *API) GetProjects(w http.ResponseWriter, r *http.Request) {

	var res Result
	projects, err := api.db.GetProjects()

	if err != nil {
		res = CodeToResult[CodeDatabaseError]
		log.Errorf("Could not retrieve project : %+v", err)
		w.WriteHeader(CodeToResult[CodeDatabaseError].HTTPCode)
		json.NewEncoder(w).Encode(res)
		return
	}

	res = CodeToResult[CodeOK]

	// delete sub field info
	for i := range projects {
		projects[i].IPs = nil
	}
	res.Data = projects

	json.NewEncoder(w).Encode(res)
}

/*
GetProject return this template
  "data": {
    "name": "FirstProject",
    "updated_at": 1520122127
  }
*/
func (api *API) GetProject(w http.ResponseWriter, r *http.Request) {
	var res Result
	vars := mux.Vars(r)

	log.Debugf("Requesting project : %s", vars["name"])
	p, err := api.db.GetProject(vars["name"])

	//TOFIX
	if err != nil && err.Error() == "not found" {
		res = CodeToResult[CodeNotFound]
		res.Message = "Project not found"

		log.Warnf("Project not found %s", vars["name"])
		w.WriteHeader(CodeToResult[CodeNotFound].HTTPCode)
		json.NewEncoder(w).Encode(res)
		return
	}

	if err != nil {
		res = CodeToResult[CodeDatabaseError]
		log.Errorf("Could not retrieve project : %+v", err)
		w.WriteHeader(CodeToResult[CodeDatabaseError].HTTPCode)
		json.NewEncoder(w).Encode(res)
		return
	}

	// we don't want all data
	if p.IPs != nil {
		p.IPs = nil
	}

	res = CodeToResult[CodeOK]
	res.Data = p

	json.NewEncoder(w).Encode(res)
}

//GetAlgorithm return the current algorithm used by the server
func (api *API) GetAlgorithm(w http.ResponseWriter, r *http.Request) {
	var res Result
	res = CodeToResult[CodeOK]
	res.Data = api.Server.Session.Algo.Name()
	json.NewEncoder(w).Encode(res)
}

func (api *API) ChangeAlgorithm(w http.ResponseWriter, r *http.Request) {
	var res Result
	var algorithm string

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&algorithm)

	if err != nil {
		res = CodeToResult[CodeCouldNotDecodeJSON]
		w.WriteHeader(CodeToResult[CodeCouldNotDecodeJSON].HTTPCode)
		json.NewEncoder(w).Encode(res)
		return
	}
	defer r.Body.Close()

	if err != nil {
		res = CodeToResult[CodeInvalidInput]
		w.WriteHeader(CodeToResult[CodeCouldNotDecodeJSON].HTTPCode)
		json.NewEncoder(w).Encode(res)
		return
	}
}

/*
GetIPsByProjectName return this template
  "data": [
    "10.0.0.1",
    "10.0.0.12",
    "10.20.3.4"
  ]
*/
func (api *API) GetIPsByProjectName(w http.ResponseWriter, r *http.Request) {
	var res Result

	vars := mux.Vars(r)
	name := vars["name"]

	// calling the private function !
	ips, err := api.db.GetIPs(name)

	// Database error
	if err != nil {
		log.Errorf("Error in selecting projects %s", err.Error())

		res = CodeToResult[CodeDatabaseError]
		res.Message += "[error in selecting project IPs]"

		w.WriteHeader(CodeToResult[CodeDatabaseError].HTTPCode)
		json.NewEncoder(w).Encode(res)
		return
	}

	log.Debugf("IPs : %+v", ips)

	// convert [{Value: "1.1.1.1"},...] to ["1.1.1.1",...]
	var data []string
	for _, ip := range ips {
		data = append(data, ip.Value)
	}

	// Not found
	if len(data) == 0 {
		log.Debugf("Project %s not found", name)
		res = CodeToResult[CodeNotFound]
		res.Message = "No IP found"

		w.WriteHeader(CodeToResult[CodeNotFound].HTTPCode)
		json.NewEncoder(w).Encode(res)
		return
	}

	res = CodeToResult[CodeOK]
	res.Data = data
	json.NewEncoder(w).Encode(res)
}

/*
GetPortsByIP return this template
  "data": [
    {
    "number": 22
    "protocol": "tcp"
    "status": "open"
    "banner": "OpenSSH..."
    "type": "ssh"
    },
    {
      ...
    }
  ]
*/
func (api *API) GetPortsByIP(w http.ResponseWriter, r *http.Request) {
	var res Result

	vars := mux.Vars(r)
	name := vars["name"]
	ip := vars["ip"]
	protocol := vars["protocol"]

	if protocol != "" {
		log.Debugf("name : %s, ip : %s, protocol : %s", name, ip, protocol)
		res = CodeToResult[CodeNotImplementedYet]

		w.WriteHeader(CodeToResult[CodeNotImplementedYet].HTTPCode)
		json.NewEncoder(w).Encode(res)
		return
	}

	ports, err := api.db.GetPorts(name, ip)

	if err != nil {
		log.Debugf("Error : %s", err)
		res = CodeToResult[CodeDatabaseError]
		json.NewEncoder(w).Encode(res)
		return
	}
	log.Debugf("ports : %s", ports)
}

/*
GetURIByPort return this template
  "data": [

  ]
}
*/
func (api *API) GetURIByPort(w http.ResponseWriter, r *http.Request) {
	//TODO
	res := CodeToResult[CodeNotImplementedYet]

	w.WriteHeader(CodeToResult[CodeNotImplementedYet].HTTPCode)
	json.NewEncoder(w).Encode(res)
}

//GetRawModuleByProject returns all the raw output for requested module.
func (api *API) GetRawModuleByProject(w http.ResponseWriter, r *http.Request) {
	//TODO
	res := CodeToResult[CodeNotImplementedYet]
	w.WriteHeader(CodeToResult[CodeNotImplementedYet].HTTPCode)
	json.NewEncoder(w).Encode(res)
}

/*
GetRoutesByIP returns all the routes info following this template :
  "data": [
    {
      "Source": "1.2.3.4",
      "Destination": "4.3.2.1",
      "Hops": {
        "IP" : "127.0.0.1",
        "Max": 0.123,
        "Min": 0.1,
        "Avg": 0.11
      }
    },
    ...
    ]
*/
func (api *API) GetRoutesByIP(w http.ResponseWriter, r *http.Request) {
	//TODO
	res := CodeToResult[CodeNotImplementedYet]
	w.WriteHeader(CodeToResult[CodeNotImplementedYet].HTTPCode)
	json.NewEncoder(w).Encode(res)
}

/*
CreateProject return this template after creating the new project
  "data": "ProjectName"
*/
func (api *API) CreateProject(w http.ResponseWriter, r *http.Request) {
	var project models.Project
	var res Result
	fmt.Println(r)
	decoder := json.NewDecoder(r.Body)
	fmt.Println("decoder : ", decoder)

	err := decoder.Decode(&project)
	if err != nil {
		log.Fatalf("Could not decode provided json : %+v", err)
		res = CodeToResult[CodeCouldNotDecodeJSON]
		w.WriteHeader(CodeToResult[CodeCouldNotDecodeJSON].HTTPCode)
		json.NewEncoder(w).Encode(res)
		return
	}

	log.Debugf("JSON input : %+v", project)
	defer r.Body.Close()

	//Create project in DBk
	api.db.CreateOrUpdateProject(project)

	res = CodeToResult[CodeOK]
	res.Message = "Command Sent"
	json.NewEncoder(w).Encode(res)
}

//RunModules runs every enabled modules
func (api *API) RunModules(w http.ResponseWriter, r *http.Request) {
	var inputs []modules.Input
	var res Result

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&inputs)
	if err != nil {
		log.Debugf("Could not decode provided json : %+v", err)

		res = CodeToResult[CodeCouldNotDecodeJSON]
		w.WriteHeader(CodeToResult[CodeCouldNotDecodeJSON].HTTPCode)
		json.NewEncoder(w).Encode(res)
		return
	}
	log.Debugf("JSON input : %+v", inputs)
	defer r.Body.Close()

	/*
	* TODO
	* Implements load balancing betweens node
	* Remove duplications
	* 	- maybe each module should look in the database and check if it has been already done
	* 	- Scan expiration ? re-runable script ? only re run if not in the same area / ip range ?
	 */

	for _, module := range api.Session.ModulesEnabled {
		moduleName := strings.ToLower(module.Name())
		cmd := communication.Command{Name: moduleName, Options: inputs}
		log.Debugf("RunModule for cmd : %+v", cmd)

		err = api.Server.SendCmd(cmd)
		if err != nil {
			res = CodeToResult[CodeNotImplementedYet]

			w.WriteHeader(CodeToResult[CodeNotImplementedYet].HTTPCode)
			json.NewEncoder(w).Encode(res)
			return
		}
	}

	res = CodeToResult[CodeOK]
	res.Message = "Command sent"
	json.NewEncoder(w).Encode(res)
}

/*
RunModule return this template after starting the modules
  "data": {
    nodes: [
      "1.2.3.4",
      "4.3.2.1"
    ]
  }
*/
func (api *API) RunModule(w http.ResponseWriter, r *http.Request) {

	var inputs []modules.Input
	var res Result

	vars := mux.Vars(r)
	module := vars["module"]

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&inputs)
	if err != nil {
		log.Debugf("Could not decode provided json : %+v", err)

		res = CodeToResult[CodeCouldNotDecodeJSON]

		w.WriteHeader(CodeToResult[CodeCouldNotDecodeJSON].HTTPCode)
		json.NewEncoder(w).Encode(res)
		return
	}

	log.Debugf("JSON input : %+v", inputs)
	defer r.Body.Close()

	cmd := communication.Command{Name: module, Options: inputs}

	log.Debugf("RunModule for cmd : %+v", cmd)

	err = api.Server.SendCmd(cmd)
	if err != nil {
		//TODO
		res = CodeToResult[CodeNotImplementedYet]
		w.WriteHeader(CodeToResult[CodeNotImplementedYet].HTTPCode)
		json.NewEncoder(w).Encode(res)
		return
	}

	res = CodeToResult[CodeOK]
	res.Message = "Command sent"
	json.NewEncoder(w).Encode(res)
}

/*
DeleteProject return this template after deleting the project
  "data": "ProjectName"
*/
func (api *API) DeleteProject(w http.ResponseWriter, r *http.Request) {
	//TODO
	res := CodeToResult[CodeNotImplementedYet]
	w.WriteHeader(CodeToResult[CodeNotImplementedYet].HTTPCode)
	json.NewEncoder(w).Encode(res)
}

func jsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debugf("Request URL : %s", r.RequestURI)
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
