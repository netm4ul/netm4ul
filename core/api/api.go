package api

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"

	"github.com/netm4ul/netm4ul/modules"

	"github.com/gorilla/mux"
	"github.com/netm4ul/netm4ul/core/config"
	"github.com/netm4ul/netm4ul/core/database"
	"github.com/netm4ul/netm4ul/core/server"
	"github.com/netm4ul/netm4ul/core/session"
	log "github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2/bson"
)

// Result is the standard response format
type Result struct {
	Status   string      `json:"status"`
	Code     Code        `json:"code"`
	Message  string      `json:"message,omitempty"`
	Data     interface{} `json:"data,omitempty"`
	HTTPCode int         `json:"-"` //remove HTTPCode from the json response
}

type API struct {
	// Session defines the global session for the API.
	Session *session.Session
	Server  *server.Server
}

type APIInfo struct {
	Port     uint16          `json:"port,omitempty"`
	Versions config.Versions `json:"versions"`
}

//Metadata of the current system (node, api, database)
type Metadata struct {
	Nodes   map[string]config.Node `json:"nodes"`
	APIInfo APIInfo                `json:"api"`
}

// CreateAPI : Initialise the infinite server loop on the master node
func CreateAPI(s *session.Session, server *server.Server) *API {
	api := API{Session: s, Server: server}
	api.Start()
	return &api
}

//Start the API and route endpoints to functions
func (api *API) Start() {

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
	router.HandleFunc(prefix+"/projects/{name}/ips", api.GetIPsByProjectName).Methods("GET")
	router.HandleFunc(prefix+"/projects/{name}/ips/{ip}/ports", api.GetPortsByIP).Methods("GET")            // We don't need to go deeper. Get all ports at once
	router.HandleFunc(prefix+"/projects/{name}/ips/{ip}/ports/{protocol}", api.GetPortsByIP).Methods("GET") // get only one protocol result (tcp, udp). Same GetPortsByIP function
	router.HandleFunc(prefix+"/projects/{name}/ips/{ip}/ports/{protocol}/{port}/directories", api.GetDirectoryByPort).Methods("GET")
	router.HandleFunc(prefix+"/projects/{name}/ips/{ip}/routes", api.GetRoutesByIP).Methods("GET")
	router.HandleFunc(prefix+"/projects/{name}/raw/{module}", api.GetRawModuleByProject).Methods("GET")

	// POST
	router.HandleFunc(prefix+"/projects", api.CreateProject).Methods("POST")
	router.HandleFunc(prefix+"/projects/{name}/run", api.RunModules).Methods("POST")
	router.HandleFunc(prefix+"/projects/{name}/run/{module}", api.RunModule).Methods("POST")

	// DELETE
	router.HandleFunc(prefix+"/projects/{name}", api.DeleteProject).Methods("DELETE")

	log.Fatal(http.ListenAndServe(ipport, router))
}

//GetIndex returns a link to the documentation on the root path
func (api *API) GetIndex(w http.ResponseWriter, r *http.Request) {

	apiInfo := APIInfo{Port: api.Session.Config.API.Port, Versions: api.Session.Config.Versions}
	d := Metadata{APIInfo: apiInfo, Nodes: api.Server.Session.Config.Nodes}

	res := CodeToResult[CodeOK]
	res.Data = d
	res.Message = "Documentation available at https://github.com/netm4ul/netm4ul"
	json.NewEncoder(w).Encode(res)
}

//GetProjects return this template
/*
{
  "status": "success",
  "code": CodeOK, // real value in /core/api/codes.go
  "data": [
    {
      "name": "FirstProject"
    }
  ]
}
*/
func (api *API) GetProjects(w http.ResponseWriter, r *http.Request) {
	sessionMgo := database.Connect()
	p := database.GetProjects(sessionMgo)

	res := CodeToResult[CodeOK]
	res.Data = p

	json.NewEncoder(w).Encode(res)
}

//GetProject return this template
/*
{
  "status": "success",
  "code": CodeOK, // real value in /core/api/codes.go
  "data": {
    "name": "FirstProject",
    "updated_at": 1520122127
  }
}
*/
func (api *API) GetProject(w http.ResponseWriter, r *http.Request) {
	var res Result
	vars := mux.Vars(r)
	sessionMgo := database.Connect()

	log.Debugf("Requesting project : %s", vars["name"])
	p := database.GetProjectByName(sessionMgo, vars["name"])

	// TODO : use real data
	p.IPs = append(p.IPs, database.IP{
		ID:    bson.NewObjectId(),
		Value: net.ParseIP("127.0.0.1"),
		Ports: []database.Port{
			database.Port{Number: 53, Banner: "Bind9", Status: "open"},
		},
	})

	if p.Name == "" {
		res = CodeToResult[CodeNotFound]
		res.Message = "Project not found"

		w.WriteHeader(CodeToResult[CodeNotFound].HTTPCode)
		json.NewEncoder(w).Encode(res)
		return
	}

	res = CodeToResult[CodeOK]
	res.Data = p

	json.NewEncoder(w).Encode(res)

}

//GetIPsByProjectName return this template
/*
{
  "status": "success",
  "code": CodeOK, // real value in /core/api/codes.go
  "data": [
	  "10.0.0.1",
	  "10.0.0.12",
	  "10.20.3.4"
  ]
}
*/
func (api *API) GetIPsByProjectName(w http.ResponseWriter, r *http.Request) {

	var res Result
	var ips []database.IP

	vars := mux.Vars(r)
	name := vars["name"]
	sessionMgo := database.Connect()
	dbCollection := api.Session.Config.Database.Database

	err := sessionMgo.DB(dbCollection).C("projects").Find(bson.M{"Name": name}).All(&ips)
	if err != nil {
		log.Errorf("Error in selecting projects %s", err.Error())

		res = CodeToResult[CodeDatabaseError]
		res.Message += "[error in selecting project IPs]"

		w.WriteHeader(CodeToResult[CodeDatabaseError].HTTPCode)
		json.NewEncoder(w).Encode(res)
		return
	}

	if len(ips) == 1 && ips[0].Value == nil {
		log.Debugf("Project %s not found", name)
		res = CodeToResult[CodeNotFound]
		res.Message = "No IP found"

		w.WriteHeader(CodeToResult[CodeNotFound].HTTPCode)
		json.NewEncoder(w).Encode(res)
		return
	}

	res = CodeToResult[CodeOK]
	res.Data = ips
	json.NewEncoder(w).Encode(res)
}

//GetPortsByIP return this template
/*
{
  "status": "success",
  "code": CodeOK, // real value in /core/api/codes.go
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
}
*/
func (api *API) GetPortsByIP(w http.ResponseWriter, r *http.Request) {
	//TODO
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

	log.Debugf("name : %s, ip : %s", name, ip)

	res = CodeToResult[CodeNotImplementedYet]
	json.NewEncoder(w).Encode(res)
}

//GetDirectoryByPort return this template
/*
{
  "status": "success",
  "code": CodeOK, // real value in /core/api/codes.go
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
}
*/
func (api *API) GetDirectoryByPort(w http.ResponseWriter, r *http.Request) {
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

//GetRoutesByIP returns all the routes info following this template :
/*
{
	"status": "success",
	"code": CodeOK, // real value in /core/api/codes.go
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
	}
*/
func (api *API) GetRoutesByIP(w http.ResponseWriter, r *http.Request) {
	//TODO
	res := CodeToResult[CodeNotImplementedYet]
	w.WriteHeader(CodeToResult[CodeNotImplementedYet].HTTPCode)
	json.NewEncoder(w).Encode(res)
}

//CreateProject return this template after creating the new project
/*
{
	"status": "success",
	"code": CodeOK, // real value in /core/api/codes.go
	"data": "ProjectName"
}
*/
func (api *API) CreateProject(w http.ResponseWriter, r *http.Request) {
	//TODO
	res := CodeToResult[CodeNotImplementedYet]

	w.WriteHeader(CodeToResult[CodeNotImplementedYet].HTTPCode)
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
		cmd := server.Command{Name: moduleName, Options: inputs}
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

//RunModule return this template after starting the modules
/*
{
	"status": "success",
	"code": CodeOK, // real value in /core/api/codes.go
	"data": {
		nodes: [
			"1.2.3.4",
			"4.3.2.1"
		]
	}
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

	cmd := server.Command{Name: module, Options: inputs}

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

//DeleteProject return this template after deleting the project
/*
{
	"status": "success",
	"code": CodeOK, // real value in /core/api/codes.go
	"data": "ProjectName"
}
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
