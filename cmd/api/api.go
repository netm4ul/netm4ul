package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/netm4ul/netm4ul/cmd/config"
	"github.com/netm4ul/netm4ul/cmd/server/database"
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

	// POST
	router.HandleFunc(APIEndpoint+"/projects", CreateProject).Methods("POST")

	// DELETE
	router.HandleFunc(APIEndpoint+"/projects/{name}", DeleteProject).Methods("DELETE")
	log.Fatal(http.ListenAndServe(ipport, router))
}

func GetIndex(w http.ResponseWriter, r *http.Request) {
}

func GetProjects(w http.ResponseWriter, r *http.Request) {
	session := database.Connect()
	p := database.GetProjects(session)
	fmt.Println(p)
	json.NewEncoder(w).Encode(p)
}

func GetProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	session := database.Connect()
	fmt.Println("Requestion project : ", vars["name"])
	p := database.GetProjectByName(session, vars["name"])
	if p.Name != "" {
		res := Result{Status: "success", Data: p}
		json.NewEncoder(w).Encode(res)
	} else {
		notFound := Result{Status: "error", Message: "Project not found"}
		json.NewEncoder(w).Encode(notFound)
	}
}

func CreateProject(w http.ResponseWriter, r *http.Request) {

}

func DeleteProject(w http.ResponseWriter, r *http.Request) {

}

func jsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.RequestURI)
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
