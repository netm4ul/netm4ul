package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/netm4ul/netm4ul/cmd/config"
	"github.com/netm4ul/netm4ul/cmd/server/database"
	"gopkg.in/mgo.v2/bson"
)

const (
	// APIVersion is the string representation of the api version
	APIVersion = "v1"
	// APIEndpoint represents the path of the api
	APIEndpoint = "/api/" + APIVersion
)

// Result is the standard response format
type Result struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

//Start the API and route endpoints to functions
func Start(ipport string, conf *config.ConfigToml) {

	log.Println("API Listenning : ", ipport)
	router := mux.NewRouter()

	// Add content-type json header !
	router.Use(jsonMiddleware)

	// GET
	router.HandleFunc("/", GetIndex).Methods("GET")
	router.HandleFunc(APIEndpoint+"/projects", GetProjects).Methods("GET")
	router.HandleFunc(APIEndpoint+"/projects/{name}", GetProject).Methods("GET")
	router.HandleFunc(APIEndpoint+"/projects/{name}/ips", GetIPsByProjectName).Methods("GET")
	router.HandleFunc(APIEndpoint+"/projects/{name}/ips/{ip}/ports", GetPortsByIP).Methods("GET") // We don't need to go deeper. Get all ports at once
	router.HandleFunc(APIEndpoint+"/projects/{name}/ips/{ip}/ports/{port}/directories", GetDirectoryByPort).Methods("GET")
	router.HandleFunc(APIEndpoint+"/projects/{name}/ips/{ip}/routes", GetRoutesByIP).Methods("GET")
	router.HandleFunc(APIEndpoint+"/projects/{name}/raw/{module}", GetRawModuleByProject).Methods("GET")

	// POST
	router.HandleFunc(APIEndpoint+"/projects", CreateProject).Methods("POST")

	// DELETE
	router.HandleFunc(APIEndpoint+"/projects/{name}", DeleteProject).Methods("DELETE")
	log.Fatal(http.ListenAndServe(ipport, router))
}

//GetIndex returns a link to the documentation on the root path
func GetIndex(w http.ResponseWriter, r *http.Request) {
	res := Result{Status: "success", Message: "Documentation available at https://github.com/netm4ul/netm4ul"}
	json.NewEncoder(w).Encode(res)
}

//GetProjects return this template
/*
{
  "status": "success",
  "data": [
    {
      "name": "FirstProject"
    }
  ]
}
*/
func GetProjects(w http.ResponseWriter, r *http.Request) {
	session := database.Connect()
	p := database.GetProjects(session)
	res := Result{Status: "success", Data: p}
	json.NewEncoder(w).Encode(res)
}

//GetProject return this template
/*
{
  "status": "success",
  "data": {
    "name": "FirstProject",
    "updated_at": 1520122127
  }
}
*/
func GetProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	session := database.Connect()
	fmt.Println("Requestion project : ", vars["name"])
	p := database.GetProjectByName(session, vars["name"])
	if p.Name != "" {
		res := Result{Status: "success", Data: p}
		json.NewEncoder(w).Encode(res)
		return
	}
	notFound := Result{Status: "error", Message: "Project not found"}
	json.NewEncoder(w).Encode(notFound)
}

//GetIPsByProjectName return this template
/*
{
  "status": "success",
  "data": [
	  "10.0.0.1",
	  "10.0.0.12",
	  "10.20.3.4"
  ]
}
*/
func GetIPsByProjectName(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	name := vars["name"]
	session := database.Connect()

	var ips []database.IP

	err := session.DB(database.DBname).C("projects").Find(bson.M{"Name": name}).All(&ips)
	if err != nil {
		log.Println("Error in selecting projects", err)
		res := Result{Status: "error", Message: "Error in selecting project IPs"}
		json.NewEncoder(w).Encode(res)
		return
	}

	if len(ips) == 1 && ips[0].Value == nil {
		res := Result{Status: "success", Data: []string{}, Message: "No IP found"}
		json.NewEncoder(w).Encode(res)
		return
	}
	res := Result{Status: "success", Data: ips}
	json.NewEncoder(w).Encode(res)
}

//GetPortsByIP return this template
/*
{
  "status": "success",
  "data": [
	  {
		"number": 22
		"protocol": "tcp"
		"status": "open"
		"banner": "OpenSSH..."
		"type": "ssh"
	  },
	  {
		  ...
	  }
  ]
}
*/
func GetPortsByIP(w http.ResponseWriter, r *http.Request) {
	//TODO
	vars := mux.Vars(r)
	name := vars["name"]
	ip := vars["ip"]
	log.Println("name :", name, "ip :", ip)

	res := Result{Status: "error", Message: "Not implemented yet"}
	json.NewEncoder(w).Encode(res)
}

//GetDirectoryByPort return this template
/*
{
  "status": "success",
  "data": [
	  {
		"number": 22
		"protocol": "tcp"
		"status": "open"
		"banner": "OpenSSH..."
		"type": "ssh"
	  },
	  {
		  ...
	  }
  ]
}
*/
func GetDirectoryByPort(w http.ResponseWriter, r *http.Request) {
	//TODO
	res := Result{Status: "error", Message: "Not implemented yet"}
	json.NewEncoder(w).Encode(res)
}

//GetRawModuleByProject returns all the raw output for requested module.
func GetRawModuleByProject(w http.ResponseWriter, r *http.Request) {
	//TODO
	res := Result{Status: "error", Message: "Not implemented yet"}
	json.NewEncoder(w).Encode(res)
}

//GetRoutesByIP returns all the routes info following this template :
/*
{
	"status": "success",
	"data": [
		{
			"Source": "1.2.3.4",
			"Destination": "4.3.2.1",
			"Hops": {
				"IP" : "127.0.0.1",
				"Max": 0.123,
				"Min": 0.1,
				"Avg": 0.11
			}
		},
		...
	]
}
*/
func GetRoutesByIP(w http.ResponseWriter, r *http.Request) {
	//TODO
	res := Result{Status: "error", Message: "Not implemented yet"}
	json.NewEncoder(w).Encode(res)
}

//CreateProject return this template after creating the new project
/*
{
	"status": "success",
	"data": "ProjectName"
}
*/
func CreateProject(w http.ResponseWriter, r *http.Request) {
	//TODO
	res := Result{Status: "error", Message: "Not implemented yet"}
	json.NewEncoder(w).Encode(res)
}

//DeleteProject return this template after deleting the project
/*
{
	"status": "success",
	"data": "ProjectName"
}
*/
func DeleteProject(w http.ResponseWriter, r *http.Request) {
	//TODO
	res := Result{Status: "error", Message: "Not implemented yet"}
	json.NewEncoder(w).Encode(res)
}

func jsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.RequestURI)
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
