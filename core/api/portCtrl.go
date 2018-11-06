package api

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/netm4ul/netm4ul/core/database/models"
	log "github.com/sirupsen/logrus"
)

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

func (api *API) PostPortsByIP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var ip string
	var err error
	vars := mux.Vars(r)
	pathUnescapeErr := 0
	if ip, err = url.PathUnescape(vars["ip"]); err != nil {
		pathUnescapeErr++
	}
	if pathUnescapeErr != 0 {
		sendInvalidArgument(w)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var port models.Port
	err = decoder.Decode(&port)
	if err != nil {
		sendInvalidArgument(w)
		return
	}
	log.Println(port)
	err = api.db.CreateOrUpdatePort(api.Session.Config.Project.Name, ip, port)
	if err != nil {
		log.Errorf("Database error : %s", err)
		sendDatabaseError(w)
		return
	}
	sendDefaultValue(w, CodeOK)
}
