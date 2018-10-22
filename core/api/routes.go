package api

import "github.com/gorilla/mux"

//Routes is responsible for seting up all the handler function for the API
func (api *API) Routes() {
	api.Router = mux.NewRouter()
	// Add content-type json header !
	api.Router.Use(api.jsonMiddleware)
	api.Router.Use(api.authMiddleware)

	// GET
	api.Router.HandleFunc(api.Prefix+"/", api.GetIndex).Methods("GET")

	api.Router.HandleFunc(api.Prefix+"/users/{name}", api.GetUser).Methods("GET")

	api.Router.HandleFunc(api.Prefix+"/projects", api.GetProjects).Methods("GET")
	api.Router.HandleFunc(api.Prefix+"/projects/{name}", api.GetProject).Methods("GET")
	api.Router.HandleFunc(api.Prefix+"/projects/{name}/algorithm", api.GetAlgorithm).Methods("GET")
	api.Router.HandleFunc(api.Prefix+"/projects/{name}/ips", api.GetIPsByProjectName).Methods("GET")
	api.Router.HandleFunc(api.Prefix+"/projects/{name}/ips/{ip}/ports", api.GetPortsByIP).Methods("GET")            // We don't need to go deeper. Get all ports at once
	api.Router.HandleFunc(api.Prefix+"/projects/{name}/ips/{ip}/ports/{protocol}", api.GetPortsByIP).Methods("GET") // get only one protocol result (tcp, udp). Same GetPortsByIP function
	api.Router.HandleFunc(api.Prefix+"/projects/{name}/ips/{ip}/ports/{protocol}/{port}/directories", api.GetURIByPort).Methods("GET")
	api.Router.HandleFunc(api.Prefix+"/projects/{name}/ips/{ip}/routes", api.GetRoutesByIP).Methods("GET")
	api.Router.HandleFunc(api.Prefix+"/projects/{name}/raw/{module}", api.GetRawModuleByProject).Methods("GET")

	// POST
	api.Router.HandleFunc(api.Prefix+"/users/create", api.CreateUser).Methods("POST")
	api.Router.HandleFunc(api.Prefix+"/users/login", api.UserLogin).Methods("POST")
	api.Router.HandleFunc(api.Prefix+"/users/logout", api.UserLogout).Methods("POST")

	api.Router.HandleFunc(api.Prefix+"/projects", api.CreateProject).Methods("POST")
	api.Router.HandleFunc(api.Prefix+"/projects/{name}/algorithm", api.ChangeAlgorithm).Methods("POST")
	api.Router.HandleFunc(api.Prefix+"/projects/{name}/run", api.RunModules).Methods("POST")
	api.Router.HandleFunc(api.Prefix+"/projects/{name}/run/{module}", api.RunModule).Methods("POST")

	// DELETE
	api.Router.HandleFunc(api.Prefix+"/projects/{name}", api.DeleteProject).Methods("DELETE")

}
