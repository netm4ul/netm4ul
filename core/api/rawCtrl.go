package api

import (
	"net/http"
)

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
