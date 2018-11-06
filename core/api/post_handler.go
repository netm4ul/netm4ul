package api

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/netm4ul/netm4ul/core/database/models"
	"github.com/netm4ul/netm4ul/core/loadbalancing"
	log "github.com/sirupsen/logrus"
)

//PostAlgorithm is the api endpoint handler for changing the loadbalancing algorithm
func (api *API) PostAlgorithm(w http.ResponseWriter, r *http.Request) {
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

func (api *API) PostProject(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	var project models.Project
	err := decoder.Decode(&project)
	if err != nil {
		sendInvalidArgument(w)
		return
	}
	log.Println(project)
	err = api.db.CreateOrUpdateProject(project)
	if err != nil {
		log.Errorf("Database error : %s", err)
		sendDatabaseError(w)
		return
	}
	sendDefaultValue(w, CodeOK)
}

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

func (api *API) PostDomain(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	var domain models.Domain
	err := decoder.Decode(&domain)
	if err != nil {
		sendInvalidArgument(w)
		return
	}
	log.Println(domain)
	err = api.db.CreateOrUpdateDomain(api.Session.Config.Project.Name, domain)
	if err != nil {
		log.Errorf("Database error : %s", err)
		sendDatabaseError(w)
		return
	}
	sendDefaultValue(w, CodeOK)
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
