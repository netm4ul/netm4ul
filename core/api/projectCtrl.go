package api

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/netm4ul/netm4ul/core/database/models"
	log "github.com/sirupsen/logrus"
)

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

func (api *API) PostProject(w http.ResponseWriter, r *http.Request) {
	var res Result
	var project models.Project

	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&project)
	if err != nil {
		sendInvalidArgument(w)
		return
	}
	log.Debugf("PostProject : %+v\n", project)
	err = api.db.CreateOrUpdateProject(project)
	if err != nil {
		log.Errorf("Database error : %s", err)
		sendDatabaseError(w)
		return
	}

	res = CodeToResult[CodeOK]
	res.Data = project
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
