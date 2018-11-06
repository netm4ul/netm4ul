package api

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/netm4ul/netm4ul/core/database/models"
	log "github.com/sirupsen/logrus"
)

//GetIndex returns a link to the documentation on the root path
func (api *API) GetIndex(w http.ResponseWriter, r *http.Request) {

	info := Info{Port: api.Session.Config.API.Port, Versions: Version}
	d := Metadata{Info: info}

	res := CodeToResult[CodeOK]
	res.Data = d
	res.Message = "Documentation available at https://github.com/netm4ul/netm4ul"
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
	var err error
	var project string

	vars := mux.Vars(r)
	pathUnescapeErr := 0
	if project, err = url.PathUnescape(vars["name"]); err != nil {
		pathUnescapeErr++
	}

	if pathUnescapeErr != 0 {
		sendInvalidArgument(w)
		return
	}

	log.Debugf("Requesting project : %s", project)
	p, err := api.db.GetProject(project)

	if err == models.ErrNotFound {
		res = CodeToResult[CodeNotFound]
		res.Message = "Project not found"

		log.Warnf("Project not found %s", project)
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

//GetAlgorithm return the current algorithm used by the server
func (api *API) GetAlgorithm(w http.ResponseWriter, r *http.Request) {
	var res Result
	res = CodeToResult[CodeOK]
	res.Data = api.Server.Session.Algo.Name()
	w.WriteHeader(res.HTTPCode)
	json.NewEncoder(w).Encode(res)
}

func (api *API) GetDomain(w http.ResponseWriter, r *http.Request) {

	var res Result
	var err error
	var project string
	var domainName string

	vars := mux.Vars(r)
	pathUnescapeErr := 0
	if project, err = url.PathUnescape(vars["name"]); err != nil {
		pathUnescapeErr++
	}
	if domainName, err = url.PathUnescape(vars["domain"]); err != nil {
		pathUnescapeErr++
	}

	if pathUnescapeErr != 0 {
		sendInvalidArgument(w)
		return
	}

	domains, err := api.db.GetDomain(project, domainName)
	if err != nil {
		res = CodeToResult[CodeDatabaseError]
		log.Errorf("Could not retrieve project : %+v", err)
		w.WriteHeader(CodeToResult[CodeDatabaseError].HTTPCode)
		json.NewEncoder(w).Encode(res)
		return
	}

	res = CodeToResult[CodeOK]
	res.Data = domains
	w.WriteHeader(res.HTTPCode)
	json.NewEncoder(w).Encode(res)
}

func (api *API) GetDomains(w http.ResponseWriter, r *http.Request) {

	var res Result
	var err error
	var project string

	vars := mux.Vars(r)
	pathUnescapeErr := 0
	if project, err = url.PathUnescape(vars["name"]); err != nil {
		pathUnescapeErr++
	}

	if pathUnescapeErr != 0 {
		sendInvalidArgument(w)
		return
	}

	domains, err := api.db.GetDomains(project)
	if err != nil {
		res = CodeToResult[CodeDatabaseError]
		log.Errorf("Could not retrieve project : %+v", err)
		w.WriteHeader(CodeToResult[CodeDatabaseError].HTTPCode)
		json.NewEncoder(w).Encode(res)
		return
	}

	res = CodeToResult[CodeOK]
	res.Data = domains
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
	var err error
	var project string

	vars := mux.Vars(r)
	pathUnescapeErr := 0
	if project, err = url.PathUnescape(vars["name"]); err != nil {
		pathUnescapeErr++
	}

	if pathUnescapeErr != 0 {
		sendInvalidArgument(w)
		return
	}

	ips, err := api.db.GetIPs(project)
	if err == models.ErrNotFound {
		log.Debugf("Ip for project %s not found", project)
		res = CodeToResult[CodeNotFound]
		res.Message = "No IP found"

		w.WriteHeader(res.HTTPCode)
		json.NewEncoder(w).Encode(res)
		return
	}

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

	res = CodeToResult[CodeOK]
	res.Data = ips
	w.WriteHeader(res.HTTPCode)
	json.NewEncoder(w).Encode(res)
}

// GetPortsByIP return all the ports for a given IP (and project)
func (api *API) GetPortsByIP(w http.ResponseWriter, r *http.Request) {
	var res Result
	var err error
	var project, ip string
	vars := mux.Vars(r)
	pathUnescapeErr := 0

	if project, err = url.PathUnescape(vars["name"]); err != nil {
		pathUnescapeErr++
	}
	if ip, err = url.PathUnescape(vars["ip"]); err != nil {
		pathUnescapeErr++
	}

	if pathUnescapeErr != 0 {
		sendInvalidArgument(w)
		return
	}
	protocol := r.FormValue("protocol")

	if protocol != "" {
		log.Debugf("project : %s, ip : %s, protocol : %s", project, ip, protocol)
		sendDefaultValue(w, CodeNotImplementedYet)
		return
	}

	ports, err := api.db.GetPorts(project, ip)
	if err != nil {
		sendDatabaseError(w)
		return
	}

	log.Debugf("ports : %+v", ports)
	if len(ports) == 0 {
		log.Debugf("No port for project %s found", project)
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
	var err error
	var project, ip, port string
	vars := mux.Vars(r)
	pathUnescapeErr := 0

	if project, err = url.PathUnescape(vars["name"]); err != nil {
		pathUnescapeErr++
	}
	if ip, err = url.PathUnescape(vars["ip"]); err != nil {
		pathUnescapeErr++
	}
	if port, err = url.PathUnescape(vars["port"]); err != nil {
		pathUnescapeErr++
	}

	if pathUnescapeErr != 0 {
		sendInvalidArgument(w)
		return
	}
	protocol := r.FormValue("protocol")

	if protocol != "" {
		log.Debugf("project : %s, ip : %s, protocol : %s", project, ip, protocol)
		sendDefaultValue(w, CodeNotImplementedYet)
		return
	}

	dbport, err := api.db.GetPort(project, ip, port)
	//Check if the port exist
	if err == models.ErrNotFound {
		log.Debugf("No port for project %s, ip %s and port %s found", project, ip, port)
		res = CodeToResult[CodeNotFound]
		res.Message = "No port found"

		w.WriteHeader(res.HTTPCode)
		json.NewEncoder(w).Encode(res)
		return
	}

	if err != nil {
		sendDatabaseError(w)
		return
	}

	res = CodeToResult[CodeOK]
	res.Data = dbport
	w.WriteHeader(res.HTTPCode)
	json.NewEncoder(w).Encode(res)
}

func (api *API) GetURIsByPort(w http.ResponseWriter, r *http.Request) {
	var res Result
	var err error
	var project, ip, port string

	vars := mux.Vars(r)
	pathUnescapeErr := 0

	if project, err = url.PathUnescape(vars["name"]); err != nil {
		pathUnescapeErr++
	}
	if ip, err = url.PathUnescape(vars["ip"]); err != nil {
		pathUnescapeErr++
	}
	if port, err = url.PathUnescape(vars["port"]); err != nil {
		pathUnescapeErr++
	}

	if pathUnescapeErr != 0 {
		sendInvalidArgument(w)
		return
	}
	protocol := r.FormValue("protocol")

	if protocol != "" {
		log.Debugf("project : %s, ip : %s, port : %s, protocol %s", project, ip, port, protocol)
		sendDefaultValue(w, CodeNotImplementedYet)
		return
	}

	uris, err := api.db.GetURIs(project, ip, port)
	if err == models.ErrNotFound {
		log.Debugf("No uris for project : %s, ip : %s, port : %s, protocol %s found", project, ip, port, protocol)
		res = CodeToResult[CodeNotFound]
		res.Message = "No URI found"

		w.WriteHeader(res.HTTPCode)
		json.NewEncoder(w).Encode(res)
		return
	}

	if err != nil {
		sendDatabaseError(w)
		return
	}

	log.Debugf("uris : %+v", uris)

	res = CodeToResult[CodeOK]
	res.Data = uris
	w.WriteHeader(res.HTTPCode)
	json.NewEncoder(w).Encode(res)
}

// GetURIByPort returns the specified URI information.
// The URI provided MUST be base64 encoded (and then urlEncoded)
// We are using standard base64 (and not the url-base64 for simplicity with frontend and other languages)
func (api *API) GetURIByPort(w http.ResponseWriter, r *http.Request) {
	var res Result
	var err error
	var project, ip, port, urib64 string

	vars := mux.Vars(r)
	pathUnescapeErr := 0

	if project, err = url.PathUnescape(vars["name"]); err != nil {
		pathUnescapeErr++
	}
	if ip, err = url.PathUnescape(vars["ip"]); err != nil {
		pathUnescapeErr++
	}
	if port, err = url.PathUnescape(vars["port"]); err != nil {
		pathUnescapeErr++
	}
	if urib64, err = url.PathUnescape(vars["uri"]); err != nil {
		pathUnescapeErr++
	}

	if pathUnescapeErr != 0 {
		sendInvalidArgument(w)
		return
	}

	uriBytes, err := base64.StdEncoding.DecodeString(urib64)
	if err != nil {
		log.Debugf("Could not decode b64 uri : %s", err)
		sendInvalidArgument(w)
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
		sendDatabaseError(w)
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

func (api *API) GetNode(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id := vars["id"]
	for _, node := range api.Server.Session.Nodes {
		if node.ID == id {
			res := CodeToResult[CodeOK]
			res.Data = node
			w.WriteHeader(res.HTTPCode)
			json.NewEncoder(w).Encode(res)
			return
		}
	}

	res := CodeToResult[CodeNotFound]
	res.Message = "Node not found"
	w.WriteHeader(CodeToResult[CodeNotFound].HTTPCode)
	json.NewEncoder(w).Encode(res)
	return

}

//GetNodes returns informations about all the connected (client) nodes for this server
func (api *API) GetNodes(w http.ResponseWriter, r *http.Request) {

	res := CodeToResult[CodeOK]
	res.Data = api.Server.Session.Nodes

	w.WriteHeader(res.HTTPCode)
	json.NewEncoder(w).Encode(res)
}
