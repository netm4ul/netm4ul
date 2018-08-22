package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/netm4ul/netm4ul/core/loadbalancing"

	"github.com/netm4ul/netm4ul/modules"

	"github.com/gorilla/mux"
	"github.com/netm4ul/netm4ul/core/communication"
	"github.com/netm4ul/netm4ul/core/database/models"
	log "github.com/sirupsen/logrus"
)

//Start the API and route endpoints to functions
func (api *API) Start() {

	api.IPPort = api.Session.GetAPIIPPort()
	api.Version = Version
	api.Prefix = "/api/" + api.Version

	log.Infof("API Listenning : %s, version : %s", api.IPPort, api.Version)
	log.Infof("API Endpoint : %s", api.IPPort+api.Prefix)

	api.Routes()

	log.Fatal(http.ListenAndServe(api.IPPort, api.Router))
}

//GetIndex returns a link to the documentation on the root path
func (api *API) GetIndex(w http.ResponseWriter, r *http.Request) {

	info := Info{Port: api.Session.Config.API.Port, Versions: Version}
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
		sendDefaultValue(w, CodeServerError)
		return
	}

	api.Server.Session.Algo = newAlgo
	res := CodeToResult[CodeOK]
	res.Message = "Algorithm changed to : " + algorithm
	w.WriteHeader(CodeToResult[CodeOK].HTTPCode)
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
		sendDefaultValue(w, CodeNotImplementedYet)
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
	sendDefaultValue(w, CodeNotImplementedYet)
}

//GetRawModuleByProject returns all the raw output for requested module.
func (api *API) GetRawModuleByProject(w http.ResponseWriter, r *http.Request) {
	//TODO
	sendDefaultValue(w, CodeNotImplementedYet)
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
	sendDefaultValue(w, CodeNotImplementedYet)
}

/*
CreateProject return this template after creating the new project
  "data": "ProjectName"
*/
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
	json.NewEncoder(w).Encode(res)
}

//TODO : use RunModule !

//RunModules runs every enabled modules
func (api *API) RunModules(w http.ResponseWriter, r *http.Request) {
	var inputs []modules.Input
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
	var inputs []modules.Input
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
