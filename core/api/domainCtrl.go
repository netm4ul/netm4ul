package api

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/netm4ul/netm4ul/core/database/models"
	log "github.com/sirupsen/logrus"
)

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

func (api *API) PostDomain(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	var domain models.Domain
	err := decoder.Decode(&domain)
	if err != nil {
		sendInvalidArgument(w)
		return
	}
	log.Debugf("PostDomain : %+v", domain)
	err = api.db.CreateOrUpdateDomain(api.Session.Config.Project.Name, domain)
	if err != nil {
		log.Errorf("Database error : %s", err)
		sendDatabaseError(w)
		return
	}
	sendDefaultValue(w, CodeOK)
}
