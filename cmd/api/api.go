package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/netm4ul/netm4ul/cmd/colors"
	"github.com/netm4ul/netm4ul/cmd/config"
	"github.com/netm4ul/netm4ul/cmd/server"
	"github.com/netm4ul/netm4ul/cmd/server/database"
	"gopkg.in/mgo.v2/bson"
)

var (
	// Version is the string representation of the api version
	Version string
)

const (
	// represents the path of the api
	CodeOK                = 200
	CodeNotFound          = 404
	CodeDatabaseError     = 998
	CodeNotImplementedYet = 999
)

// Result is the standard response format
type Result struct {
	Status  string      `json:"status"`
	Code    int         `json:"code"`
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

//Start the API and route endpoints to functions
func Start(ipport string, conf *config.ConfigToml) {
	Version = config.Config.Versions.Api

	log.Printf(colors.Green("API Listenning : %s, version : %s"), ipport, config.Config.Versions.Api)
	router := mux.NewRouter()

	// Add content-type json header !
	router.Use(jsonMiddleware)

	// GET
	router.HandleFunc("/", GetIndex).Methods("GET")
	router.HandleFunc("/projects", GetProjects).Methods("GET")
	router.HandleFunc("/projects/{name}", GetProject).Methods("GET")
	router.HandleFunc("/projects/{name}/ips", GetIPsByProjectName).Methods("GET")
	router.HandleFunc("/projects/{name}/ips/{ip}/ports", GetPortsByIP).Methods("GET")            // We don't need to go deeper. Get all ports at once
	router.HandleFunc("/projects/{name}/ips/{ip}/ports/{protocol}", GetPortsByIP).Methods("GET") // get only one protocol result (tcp, udp). Same GetPortsByIP function
	router.HandleFunc("/projects/{name}/ips/{ip}/ports/{protocol}/{port}/directories", GetDirectoryByPort).Methods("GET")
	router.HandleFunc("/projects/{name}/ips/{ip}/routes", GetRoutesByIP).Methods("GET")
	router.HandleFunc("/projects/{name}/raw/{module}", GetRawModuleByProject).Methods("GET")

	// POST
	router.HandleFunc("/projects", CreateProject).Methods("POST")
	router.HandleFunc("/projects/{name}/run/{module}", RunModule).Methods("POST")

	// DELETE
	router.HandleFunc("/projects/{name}", DeleteProject).Methods("DELETE")
	log.Fatal(http.ListenAndServe(ipport, router))
}

//GetIndex returns a link to the documentation on the root path
func GetIndex(w http.ResponseWriter, r *http.Request) {
	api := API{Port: config.Config.API.Port, Versions: config.Config.Versions}
	d := Metadata{API: api, Nodes: server.ConfigServer.Nodes}
	res := Result{Status: "success", Code: CodeOK, Message: "Documentation available at https://github.com/netm4ul/netm4ul", Data: d}
	json.NewEncoder(w).Encode(res)
}

//GetProjects return this template
/*
{
  "status": "success",
  "code": 200,
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
	// psend := struct{ Projects []database.Project }{Projects: p}
	res := Result{Status: "success", Code: CodeOK, Data: p}
	json.NewEncoder(w).Encode(res)
}

//GetProject return this template
/*
{
  "status": "success",
  "code": 200,
  "data": {
    "name": "FirstProject",
    "updated_at": 1520122127
  }
}
*/
func GetProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	session := database.Connect()

	if config.Config.Verbose {
		log.Printf(colors.Yellow("Requesting project : %s"), vars["name"])
	}
	p := database.GetProjectByName(session, vars["name"])
	if p.Name != "" {
		res := Result{Status: "success", Code: CodeOK, Data: p}
		json.NewEncoder(w).Encode(res)
		return
	}
	notFound := Result{Status: "error", Code: CodeNotFound, Message: "Project not found"}
	json.NewEncoder(w).Encode(notFound)
}

//GetIPsByProjectName return this template
/*
{
  "status": "success",
  "code": 200,
  "data": [
	  "10.0.0.1",
	  "10.0.0.12",
	  "10.20.3.4"
  ]
}
*/
func GetIPsByProjectName(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	name := vars["name"]
	session := database.Connect()

	var ips []database.IP

	err := session.DB(database.DBname).C("projects").Find(bson.M{"Name": name}).All(&ips)
	if err != nil {
		log.Println(colors.Red("Error in selecting projects %s"), err.Error())
		res := Result{Status: "error", Code: CodeDatabaseError, Message: "Error in selecting project IPs"}
		json.NewEncoder(w).Encode(res)
		return
	}

	if len(ips) == 1 && ips[0].Value == nil {
		if config.Config.Verbose {
			log.Printf(colors.Yellow("Project %s not found"), name)
		}
		res := Result{Status: "error", Code: CodeNotFound, Data: []string{}, Message: "No IP found"}
		json.NewEncoder(w).Encode(res)
		return
	}
	res := Result{Status: "success", Code: CodeOK, Data: ips}
	json.NewEncoder(w).Encode(res)
}

//GetPortsByIP return this template
/*
{
  "status": "success",
  "code": 200,
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
	vars := mux.Vars(r)
	name := vars["name"]
	ip := vars["ip"]
	protocol := vars["protocol"]

	if protocol != "" {
		if config.Config.Verbose {
			log.Printf(colors.Yellow("name : %s, ip : %s, protocol : %s"), name, ip, protocol)
		}
		res := Result{Status: "error", Code: CodeNotImplementedYet, Message: "Not implemented yet"}
		json.NewEncoder(w).Encode(res)
		return
	}

	if config.Config.Verbose {
		log.Printf(colors.Yellow("name : %s, ip : %s"), name, ip)
	}

	res := Result{Status: "error", Code: CodeNotImplementedYet, Message: "Not implemented yet"}
	json.NewEncoder(w).Encode(res)
}

//GetDirectoryByPort return this template
/*
{
  "status": "success",
  "code": 200,
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
	res := Result{Status: "error", Code: CodeNotImplementedYet, Message: "Not implemented yet"}
	json.NewEncoder(w).Encode(res)
}

//GetRawModuleByProject returns all the raw output for requested module.
func GetRawModuleByProject(w http.ResponseWriter, r *http.Request) {
	//TODO
	res := Result{Status: "error", Code: CodeNotImplementedYet, Message: "Not implemented yet"}
	json.NewEncoder(w).Encode(res)
}

//GetRoutesByIP returns all the routes info following this template :
/*
{
	"status": "success",
	"code": 200,
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
	res := Result{Status: "error", Code: CodeNotImplementedYet, Message: "Not implemented yet"}
	json.NewEncoder(w).Encode(res)
}

//CreateProject return this template after creating the new project
/*
{
	"status": "success",
	"code": 200,
	"data": "ProjectName"
}
*/
func CreateProject(w http.ResponseWriter, r *http.Request) {
	//TODO
	res := Result{Status: "error", Code: CodeNotImplementedYet, Message: "Not implemented yet"}
	json.NewEncoder(w).Encode(res)
}

//RunModule return this template after starting the modules
/*
{
	"status": "success",
	"code": 200,
	"data": {
		nodes: [
			"1.2.3.4",
			"4.3.2.1"
		]
	}
}
*/
func RunModule(w http.ResponseWriter, r *http.Request) {
	//TODO
	// Setup correct command (and option through parameters)

	vars := mux.Vars(r)
	// name := vars["name"]
	module := vars["module"]

	var res Result

	cmd := server.Command{Name: module, Options: []string{"option1", "option2"}}

	if config.Config.Verbose {
		log.Printf(colors.Yellow("RunModule for cmd : %+v"), cmd)
	}

	err := server.SendCmd(cmd)
	if err != nil {
		res = Result{Status: "error", Code: CodeNotImplementedYet, Message: "Not implemented yet"}
	}
	res = Result{Status: "success", Code: CodeOK, Message: "Command sent"}
	json.NewEncoder(w).Encode(res)
}

//DeleteProject return this template after deleting the project
/*
{
	"status": "success",
	"code": 200,
	"data": "ProjectName"
}
*/
func DeleteProject(w http.ResponseWriter, r *http.Request) {
	//TODO
	res := Result{Status: "error", Code: CodeNotImplementedYet, Message: "Not implemented yet"}
	json.NewEncoder(w).Encode(res)
}

func jsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if config.Config.Verbose {
			log.Printf(colors.Green("Request URL : %s"), r.RequestURI)
		}
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
