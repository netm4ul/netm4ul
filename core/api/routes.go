package api

import (
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

//Handler return a new mux router. All
func (api *API) Handler() *mux.Router {

	ipport := api.Session.GetAPIIPPort()
	version := api.Session.Config.Versions.Api
	prefix := "/api/" + version

	log.Infof("API Listenning : %s, version : %s", ipport, version)
	log.Infof("API Endpoint : %s", ipport+prefix)

	router := mux.NewRouter()

	// Add content-type json header !
	router.Use(api.jsonMiddleware)
	router.Use(api.authMiddleware)

	// GET
	router.HandleFunc(prefix+"/", api.GetIndex).Methods("GET")

	router.HandleFunc(prefix+"/users/{name}", api.GetUser).Methods("GET")

	router.HandleFunc(prefix+"/projects", api.GetProjects).Methods("GET")
	router.HandleFunc(prefix+"/projects/{name}", api.GetProject).Methods("GET")
	router.HandleFunc(prefix+"/projects/{name}/algorithm", api.GetAlgorithm).Methods("GET")
	router.HandleFunc(prefix+"/projects/{name}/ips", api.GetIPsByProjectName).Methods("GET")
	router.HandleFunc(prefix+"/projects/{name}/ips/{ip}/ports", api.GetPortsByIP).Methods("GET")            // We don't need to go deeper. Get all ports at once
	router.HandleFunc(prefix+"/projects/{name}/ips/{ip}/ports/{protocol}", api.GetPortsByIP).Methods("GET") // get only one protocol result (tcp, udp). Same GetPortsByIP function
	router.HandleFunc(prefix+"/projects/{name}/ips/{ip}/ports/{protocol}/{port}/directories", api.GetURIByPort).Methods("GET")
	router.HandleFunc(prefix+"/projects/{name}/ips/{ip}/routes", api.GetRoutesByIP).Methods("GET")
	router.HandleFunc(prefix+"/projects/{name}/raw/{module}", api.GetRawModuleByProject).Methods("GET")

	// POST
	router.HandleFunc(prefix+"/users/create", api.CreateUser).Methods("POST")
	router.HandleFunc(prefix+"/users/login", api.UserLogin).Methods("POST")
	router.HandleFunc(prefix+"/users/logout", api.UserLogout).Methods("POST")

	router.HandleFunc(prefix+"/projects", api.CreateProject).Methods("POST")
	router.HandleFunc(prefix+"/projects/{name}/algorithm", api.ChangeAlgorithm).Methods("POST")
	router.HandleFunc(prefix+"/projects/{name}/run", api.RunModules).Methods("POST")
	router.HandleFunc(prefix+"/projects/{name}/run/{module}", api.RunModule).Methods("POST")

	// DELETE
	router.HandleFunc(prefix+"/projects/{name}", api.DeleteProject).Methods("DELETE")
	return router
}
