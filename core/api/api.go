package api

import (
	"encoding/json"
	"net"
	"net/http"

	"github.com/netm4ul/netm4ul/modules"

	"github.com/gorilla/mux"
	"github.com/netm4ul/netm4ul/core/config"
	"github.com/netm4ul/netm4ul/core/database"
	"github.com/netm4ul/netm4ul/core/server"
	"github.com/netm4ul/netm4ul/core/session"
	log "github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2/bson"
)

var (
	// Version is the string representation of the api version
	Version    string
	SessionAPI *session.Session
)

// Result is the standard response format
type Result struct {
	Status  string      `json:"status"`
	Code    Code        `json:"code"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type API struct {
	Port     uint16          `json:"port,omitempty"`
	Versions config.Versions `json:"versions"`
}

//Metadata of the current system (node, api, database)
type Metadata struct {
	Nodes map[string]config.Node `json:"nodes"`
	API   API                    `json:"api"`
}

// CreateAPI : Initialise the infinite server loop on the master node
func CreateAPI(s *session.Session) {
	SessionAPI = s
	Start()
}

//Start the API and route endpoints to functions
func Start() {

	ipport := SessionAPI.GetAPIIPPort()
	Version = SessionAPI.Config.Versions.Api
	prefix := "/api/" + Version
	log.Infof("API Listenning : %s, version : %s", ipport, SessionAPI.Config.Versions.Api)
	log.Infof("API Endpoint : %s", ipport+prefix)
	router := mux.NewRouter()

	// Add content-type json header !
	router.Use(jsonMiddleware)

	// GET
	router.HandleFunc(prefix+"/", GetIndex).Methods("GET")
	router.HandleFunc(prefix+"/projects", GetProjects).Methods("GET")
	router.HandleFunc(prefix+"/projects/{name}", GetProject).Methods("GET")
	router.HandleFunc(prefix+"/projects/{name}/ips", GetIPsByProjectName).Methods("GET")
	router.HandleFunc(prefix+"/projects/{name}/ips/{ip}/ports", GetPortsByIP).Methods("GET")            // We don't need to go deeper. Get all ports at once
	router.HandleFunc(prefix+"/projects/{name}/ips/{ip}/ports/{protocol}", GetPortsByIP).Methods("GET") // get only one protocol result (tcp, udp). Same GetPortsByIP function
	router.HandleFunc(prefix+"/projects/{name}/ips/{ip}/ports/{protocol}/{port}/directories", GetDirectoryByPort).Methods("GET")
	router.HandleFunc(prefix+"/projects/{name}/ips/{ip}/routes", GetRoutesByIP).Methods("GET")
	router.HandleFunc(prefix+"/projects/{name}/raw/{module}", GetRawModuleByProject).Methods("GET")

	// POST
	router.HandleFunc(prefix+"/projects", CreateProject).Methods("POST")
	router.HandleFunc(prefix+"/projects/{name}/run/{module}", RunModule).Methods("POST")

	// DELETE
	router.HandleFunc(prefix+"/projects/{name}", DeleteProject).Methods("DELETE")

	log.Fatal(http.ListenAndServe(ipport, router))
}

//GetIndex returns a link to the documentation on the root path
func GetIndex(w http.ResponseWriter, r *http.Request) {

	api := API{Port: SessionAPI.Config.API.Port, Versions: SessionAPI.Config.Versions}
	d := Metadata{API: api, Nodes: server.SessionServer.Config.Nodes}

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
func GetProjects(w http.ResponseWriter, r *http.Request) {
	session := database.Connect()
	p := database.GetProjects(session)

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
func GetProject(w http.ResponseWriter, r *http.Request) {
	var res Result
	vars := mux.Vars(r)
	session := database.Connect()

	log.Debugf("Requesting project : %s", vars["name"])
	p := database.GetProjectByName(session, vars["name"])

	// TODO : use real data
	p.IPs = append(p.IPs, database.IP{
		Value: net.ParseIP("127.0.0.1"),
		Ports: []database.Port{
			database.Port{Number: 53, Banner: "Bind9", Status: "open"},
		},
	})

	if p.Name == "" {
		res = CodeToResult[CodeNotFound]
		res.Message = "Project not found"
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
func GetIPsByProjectName(w http.ResponseWriter, r *http.Request) {

	var res Result
	var ips []database.IP

	vars := mux.Vars(r)
	name := vars["name"]
	session := database.Connect()

	err := session.DB(database.DBname).C("projects").Find(bson.M{"Name": name}).All(&ips)
	if err != nil {
		log.Errorf("Error in selecting projects %s", err.Error())

		res = CodeToResult[CodeDatabaseError]
		res.Message += "[error in selecting project IPs]"
		json.NewEncoder(w).Encode(res)
		return
	}

	if len(ips) == 1 && ips[0].Value == nil {
		log.Debugf("Project %s not found", name)
		res = CodeToResult[CodeNotFound]
		res.Message = "No IP found"
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
func GetPortsByIP(w http.ResponseWriter, r *http.Request) {
	//TODO
	var res Result
	vars := mux.Vars(r)
	name := vars["name"]
	ip := vars["ip"]
	protocol := vars["protocol"]

	if protocol != "" {
		log.Debugf("name : %s, ip : %s, protocol : %s", name, ip, protocol)
		res = CodeToResult[CodeNotImplementedYet]
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
func GetDirectoryByPort(w http.ResponseWriter, r *http.Request) {
	//TODO
	res := CodeToResult[CodeNotImplementedYet]
	json.NewEncoder(w).Encode(res)
}

//GetRawModuleByProject returns all the raw output for requested module.
func GetRawModuleByProject(w http.ResponseWriter, r *http.Request) {
	//TODO
	res := CodeToResult[CodeNotImplementedYet]
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
func GetRoutesByIP(w http.ResponseWriter, r *http.Request) {
	//TODO
	res := CodeToResult[CodeNotImplementedYet]
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
func CreateProject(w http.ResponseWriter, r *http.Request) {
	//TODO
	res := CodeToResult[CodeNotImplementedYet]
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
func RunModule(w http.ResponseWriter, r *http.Request) {

	var inputs []modules.Input
	var res Result

	vars := mux.Vars(r)
	module := vars["module"]

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&inputs)
	if err != nil {
		log.Debugf("Could not decode provided json : %+v", err)

		res = CodeToResult[CodeCouldNotDecodeJSON]
		json.NewEncoder(w).Encode(res)
		return
	}

	log.Debugf("JSON input : %+v", inputs)
	defer r.Body.Close()

	cmd := server.Command{Name: module, Options: inputs}

	log.Debugf("RunModule for cmd : %+v", cmd)

	err = server.SendCmd(cmd, SessionAPI)
	if err != nil {
		res = CodeToResult[CodeNotImplementedYet]
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
func DeleteProject(w http.ResponseWriter, r *http.Request) {
	//TODO
	res := CodeToResult[CodeNotImplementedYet]
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
