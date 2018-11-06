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
