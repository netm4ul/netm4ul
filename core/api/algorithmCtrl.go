package api

import (
	"encoding/json"
	"net/http"

	"github.com/netm4ul/netm4ul/core/loadbalancing"
)

//GetAlgorithm return the current algorithm used by the server
func (api *API) GetAlgorithm(w http.ResponseWriter, r *http.Request) {
	var res Result
	res = CodeToResult[CodeOK]
	res.Data = api.Server.Session.Algo.Name()
	w.WriteHeader(res.HTTPCode)
	json.NewEncoder(w).Encode(res)
}

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
