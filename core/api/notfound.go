package api

import (
	"encoding/json"
	"net/http"
)

//NotFound returns a simple 404 json error messages
func (api *API) NotFound(w http.ResponseWriter, r *http.Request) {
	res := CodeToResult[CodeNotFound]
	res.Message = "This API endpoint does not exist."
	w.WriteHeader(res.HTTPCode)
	json.NewEncoder(w).Encode(res)
}
