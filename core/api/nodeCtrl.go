package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

//GetNodes returns informations about all the connected (client) nodes for this server
func (api *API) GetNodes(w http.ResponseWriter, r *http.Request) {

	res := CodeToResult[CodeOK]
	res.Data = api.Server.Session.Nodes

	w.WriteHeader(res.HTTPCode)
	json.NewEncoder(w).Encode(res)
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
