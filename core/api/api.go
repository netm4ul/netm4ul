package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"encoding/base64"

	"github.com/netm4ul/netm4ul/core/loadbalancing"
	"time"

	"github.com/gorilla/mux"
	"github.com/netm4ul/netm4ul/core/communication"
	"github.com/netm4ul/netm4ul/core/database/models"
	"github.com/netm4ul/netm4ul/core/server"
	"github.com/netm4ul/netm4ul/core/session"
	log "github.com/sirupsen/logrus"
)

var c chan os.Signal

//NewAPI is the constructor method for the HTTP API
func NewAPI(s *session.Session, server *server.Server) *API {
	api := API{
		Session: s,
		Server:  server,
		db:      server.Db,
		IPPort:  s.GetAPIIPPort(),
		Version: Version,
		Prefix:  "/api/" + Version,
	}

	api.Routes()
	api.setupSignal()
	return &api
}

func (api *API) setupSignal() {
	log.Debug("Creating signal channel to gracefully close the API.")
	c = make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)
}

//Start the API and route endpoints to functions
func (api *API) Start() {
	// timeout before forcing shutdown
	wait := time.Second * 3

	// this is from the mux documentation
	srv := &http.Server{
		Addr: api.IPPort,
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      api.Router, // Pass our instance of gorilla/mux in.
	}

	go func() {
		log.Infof("API Listenning : %s, version : %s", api.IPPort, api.Version)
		log.Infof("API Endpoint : %s", api.IPPort+api.Prefix)
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()
	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Println("shutting down")
	os.Exit(0)
}

//Shutdown is responsible for graceful shutdown of the API.
func (api *API) Shutdown() {
	c <- os.Interrupt
}

//GetIndex returns a link to the documentation on the root path
func (api *API) GetIndex(w http.ResponseWriter, r *http.Request) {

	info := Info{Port: api.Session.Config.API.Port, Versions: Version}
	d := Metadata{Info: info, Nodes: api.Server.Session.Nodes}

	res := CodeToResult[CodeOK]
	res.Data = d
	res.Message = "Documentation available at https://github.com/netm4ul/netm4ul"
	w.WriteHeader(res.HTTPCode)
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
	res.Data = projects
	w.WriteHeader(res.HTTPCode)
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

	if err == models.ErrNotFound {
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

	res = CodeToResult[CodeOK]
	res.Data = p
	w.WriteHeader(res.HTTPCode)
	json.NewEncoder(w).Encode(res)
}

//GetAlgorithm return the current algorithm used by the server
func (api *API) GetAlgorithm(w http.ResponseWriter, r *http.Request) {
	var res Result
	res = CodeToResult[CodeOK]
	res.Data = api.Server.Session.Algo.Name()
	w.WriteHeader(res.HTTPCode)
	json.NewEncoder(w).Encode(res)
}

//ChangeAlgorithm is the api endpoint handler for changing the loadbalancing algorithm
func (api *API) ChangeAlgorithm(w http.ResponseWriter, r *http.Request) {
	var algorithm string

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&algorithm)

	if err != nil {
		sendDefaultValue(w, CodeCouldNotDecodeJSON)
		return
	}
	defer r.Body.Close()

	newAlgo, err := loadbalancing.NewAlgo(algorithm)
	if err != nil {
		sendDefaultValue(w, CodeInvalidInput)
		return
	}

	api.Server.Session.Algo = newAlgo
	res := CodeToResult[CodeOK]
	res.Message = "Algorithm changed to : " + algorithm
	w.WriteHeader(res.HTTPCode)
	json.NewEncoder(w).Encode(res)
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

		w.WriteHeader(res.HTTPCode)
		json.NewEncoder(w).Encode(res)
		return
	}

	log.Debugf("IPs : %+v", ips)

	// Not found
	if len(ips) == 0 {
		log.Debugf("Ip for project %s not found", name)
		res = CodeToResult[CodeNotFound]
		res.Message = "No IP found"

		w.WriteHeader(res.HTTPCode)
		json.NewEncoder(w).Encode(res)
		return
	}

	res = CodeToResult[CodeOK]
	res.Data = ips
	w.WriteHeader(res.HTTPCode)
	json.NewEncoder(w).Encode(res)
}

// GetPortsByIP return all the ports for a given IP (and project)
func (api *API) GetPortsByIP(w http.ResponseWriter, r *http.Request) {
	var res Result

	vars := mux.Vars(r)
	name := vars["name"]
	ip := vars["ip"]
	protocol := r.FormValue("protocol")

	if protocol != "" {
		log.Debugf("name : %s, ip : %s, protocol : %s", name, ip, protocol)
		sendDefaultValue(w, CodeNotImplementedYet)
		return
	}

	ports, err := api.db.GetPorts(name, ip)
	if err != nil {
		log.Debugf("Error : %s", err)
		res = CodeToResult[CodeDatabaseError]
		w.WriteHeader(res.HTTPCode)
		json.NewEncoder(w).Encode(res)
		return
	}

	log.Debugf("ports : %+v", ports)
	if len(ports) == 0 {
		log.Debugf("No port for project %s found", name)
		res = CodeToResult[CodeNotFound]
		res.Message = "No port found"

		w.WriteHeader(res.HTTPCode)
		json.NewEncoder(w).Encode(res)
		return
	}

	res = CodeToResult[CodeOK]
	res.Data = ports
	w.WriteHeader(res.HTTPCode)
	json.NewEncoder(w).Encode(res)
}

// GetPortByIP return informations about a given port. It requires port number, ip address, and project name
// It has an optionnal "protocol" query (form) to specify the tcp/upd/... protocol
// If multiple ports are found with the same number and the protocol isn't specified, the function return an error : CodeAmbiguousRequest
func (api *API) GetPortByIP(w http.ResponseWriter, r *http.Request) {
	var res Result

	vars := mux.Vars(r)
	name := vars["name"]
	ip := vars["ip"]
	port := vars["port"]
	protocol := r.FormValue("protocol")

	if protocol != "" {
		log.Debugf("name : %s, ip : %s, protocol : %s", name, ip, protocol)
		sendDefaultValue(w, CodeNotImplementedYet)
		return
	}

	dbport, err := api.db.GetPort(name, ip, port)
	if err != nil {
		log.Debugf("Error : %s", err)
		res = CodeToResult[CodeDatabaseError]
		w.WriteHeader(res.HTTPCode)
		json.NewEncoder(w).Encode(res)
		return
	}

	//Check if the port exist
	if dbport.CreatedAt.IsZero() {
		log.Debugf("No port for project %s, ip %s and port %s found", name, ip, port)
		res = CodeToResult[CodeNotFound]
		res.Message = "No port found"

		w.WriteHeader(res.HTTPCode)
		json.NewEncoder(w).Encode(res)
		return
	}

	res = CodeToResult[CodeOK]
	res.Data = dbport
	w.WriteHeader(res.HTTPCode)
	json.NewEncoder(w).Encode(res)
}

func (api *API) GetURIsByPort(w http.ResponseWriter, r *http.Request) {
	var res Result

	vars := mux.Vars(r)
	project := vars["name"]
	ip := vars["ip"]
	port := vars["port"]
	protocol := r.FormValue("protocol")

	if protocol != "" {
		log.Debugf("project : %s, ip : %s, port : %s, protocol %s", project, ip, port, protocol)
		sendDefaultValue(w, CodeNotImplementedYet)
		return
	}

	uris, err := api.db.GetURIs(project, ip, port)
	if err == models.ErrNotFound || len(uris) == 0 {
		log.Debugf("No uris for project : %s, ip : %s, port : %s, protocol %s found", project, ip, port, protocol)
		res = CodeToResult[CodeNotFound]
		res.Message = "No URI found"

		w.WriteHeader(res.HTTPCode)
		json.NewEncoder(w).Encode(res)
		return
	}

	if err != nil {
		log.Debugf("Error : %s", err)
		res = CodeToResult[CodeDatabaseError]
		w.WriteHeader(res.HTTPCode)
		json.NewEncoder(w).Encode(res)
		return
	}

	log.Debugf("uris : %+v", uris)

	res = CodeToResult[CodeOK]
	res.Data = uris
	w.WriteHeader(res.HTTPCode)
	json.NewEncoder(w).Encode(res)
}

// GetURIByPort returns the specified URI information.
// The HTTP request *MUST* base64 encode this variable.
// (Golang doesn't parse correctly the input if there is a / (even encoded with %2F) and use it as a url path)
// (see https://github.com/golang/go/issues/21955 )
func (api *API) GetURIByPort(w http.ResponseWriter, r *http.Request) {
	var res Result
	var err error

	vars := mux.Vars(r)
	project := vars["name"]
	ip := vars["ip"]
	port := vars["port"]
	uriEncoded := vars["uri"]
	uriBytes, err := base64.StdEncoding.DecodeString(uriEncoded)
	if err != nil {
		log.Debugf("Error : %s", err)
		res = CodeToResult[CodeServerError]
		w.WriteHeader(res.HTTPCode)
		json.NewEncoder(w).Encode(res)
		return
	}
	uri := string(uriBytes)

	protocol := r.FormValue("protocol")

	if protocol != "" {
		log.Debugf("project : %s, ip : %s, port : %s, protocol %s", project, ip, port, protocol)
		sendDefaultValue(w, CodeNotImplementedYet)
		return
	}

	dburi, err := api.db.GetURI(project, ip, port, uri)
	if err == models.ErrNotFound {
		log.Debugf("No uri for project : %s, ip : %s, port : %s, protocol %s found", project, ip, port, protocol)
		res = CodeToResult[CodeNotFound]
		res.Message = "No URI found"

		w.WriteHeader(res.HTTPCode)
		json.NewEncoder(w).Encode(res)
		return
	}

	if err != nil {
		log.Debugf("Error : %s", err)
		res = CodeToResult[CodeDatabaseError]
		w.WriteHeader(res.HTTPCode)
		json.NewEncoder(w).Encode(res)
		return
	}

	res = CodeToResult[CodeOK]
	res.Data = dburi
	w.WriteHeader(res.HTTPCode)
	json.NewEncoder(w).Encode(res)
}

//GetRawsByProject returns all the raws output from a module for a specified project
func (api *API) GetRawsByProject(w http.ResponseWriter, r *http.Request) {
	//TODO
	sendDefaultValue(w, CodeNotImplementedYet)
}

//GetRawsByModule returns all the raws output from the requested module and project
func (api *API) GetRawsByModule(w http.ResponseWriter, r *http.Request) {
	//TODO
	sendDefaultValue(w, CodeNotImplementedYet)
}

// GetRoutesByIP returns all the routes informations
func (api *API) GetRoutesByIP(w http.ResponseWriter, r *http.Request) {
	//TODO
	sendDefaultValue(w, CodeNotImplementedYet)
}

// CreateProject creates a new project and return its name inside the data field
func (api *API) CreateProject(w http.ResponseWriter, r *http.Request) {
	var project models.Project
	var res Result

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&project)

	if err != nil {
		log.Debugf("Could not decode provided json : %+v", err)
		sendDefaultValue(w, CodeCouldNotDecodeJSON)
		return
	}

	log.Debugf("JSON input : %+v", project)
	defer r.Body.Close()

	//Create project in DBk
	api.db.CreateOrUpdateProject(project)

	res = CodeToResult[CodeOK]
	res.Message = "Command Sent"
	w.WriteHeader(res.HTTPCode)
	json.NewEncoder(w).Encode(res)
}

//TODO : use RunModule !

//RunModules runs every enabled modules
func (api *API) RunModules(w http.ResponseWriter, r *http.Request) {
	var inputs []communication.Input
	var res Result

	err := json.NewDecoder(r.Body).Decode(&inputs)
	if err != nil {
		log.Debugf("Could not decode provided json : %+v", err)
		sendDefaultValue(w, CodeCouldNotDecodeJSON)
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
		// send as much command as inputs
		for _, input := range inputs {
			cmd := communication.Command{Name: moduleName, Options: input}
			log.Debugf("RunModule for cmd : %+v", cmd)

			err = api.Server.SendCmd(cmd)

			// exit at first error.
			if err != nil {
				sendDefaultValue(w, CodeServerError)
				return
			}
		}
	}

	res = CodeToResult[CodeOK]
	res.Message = "Command sent"
	w.WriteHeader(res.HTTPCode)
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
	fmt.Printf("RunModule api.db : %+v", api.db)
	var inputs []communication.Input
	var res Result

	vars := mux.Vars(r)
	module := vars["module"]

	err := json.NewDecoder(r.Body).Decode(&inputs)

	if err != nil {
		log.Debugf("Could not decode provided json : %+v", err)
		sendDefaultValue(w, CodeCouldNotDecodeJSON)
		return
	}
	defer r.Body.Close()

	for _, input := range inputs {
		cmd := communication.Command{Name: module, Options: input}
		log.Debugf("RunModule for cmd : %+v", cmd)
		err = api.Server.SendCmd(cmd)
		if err != nil {
			//TODO
			sendDefaultValue(w, CodeNotImplementedYet)
			return
		}
	}

	res = CodeToResult[CodeOK]
	res.Message = "Command sent"
	w.WriteHeader(res.HTTPCode)
	json.NewEncoder(w).Encode(res)
}

/*
DeleteProject return this template after deleting the project
  "data": "ProjectName"
*/
func (api *API) DeleteProject(w http.ResponseWriter, r *http.Request) {
	//TODO
	sendDefaultValue(w, CodeNotImplementedYet)
}
