package api

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/netm4ul/netm4ul/core/database/models"
	log "github.com/sirupsen/logrus"
)

func (api *API) PostIP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	var IP models.IP
	err := decoder.Decode(&IP)
	if err != nil {
		sendInvalidArgument(w)
		return
	}
	log.Println(IP)
	err = api.db.CreateOrUpdateIP(api.Session.Config.Project.Name, IP)
	if err != nil {
		log.Errorf("Database error : %s", err)
		sendDatabaseError(w)
		return
	}
	sendDefaultValue(w, CodeOK)
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
