package api

import (
	"context"

	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/netm4ul/netm4ul/core/communication"
	"github.com/netm4ul/netm4ul/core/server"
	"github.com/netm4ul/netm4ul/core/session"
	log "github.com/sirupsen/logrus"
)

var c chan os.Signal

//NewAPI is the constructor method for the HTTP API
func NewAPI(s *session.Session, server *server.Server) *API {
	api := API{
		Session: s,
		Server:  server,
		db:      server.Db,
		IPPort:  s.GetAPIIPPort(),
		Version: Version,
		Prefix:  "/api/" + Version,
	}

	api.Routes()
	api.setupSignal()
	return &api
}

func (api *API) setupSignal() {
	log.Debug("Creating signal channel to gracefully close the API.")
	c = make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)
}

//Start the API and route endpoints to functions
func (api *API) Start() {
	// timeout before forcing shutdown
	wait := time.Second * 3

	// this is from the mux documentation
	srv := &http.Server{
		Addr: api.IPPort,
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		//TOFIX ! load from config
		Handler: handlers.CORS(
			handlers.AllowedOrigins([]string{"*"}),
			handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
			handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "X-Session-Token"}),
		)(api.Router), // Pass our instance of gorilla/mux in.
	}

	go func() {
		log.Infof("API Listenning : %s, version : %s", api.IPPort, api.Version)
		log.Infof("API Endpoint : %s", api.IPPort+api.Prefix)
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()
	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Println("shutting down")
	os.Exit(0)
}

//Shutdown is responsible for graceful shutdown of the API.
func (api *API) Shutdown() {
	c <- os.Interrupt
}

//TODO : use RunModule !

//RunModules runs every enabled modules
func (api *API) RunModules(w http.ResponseWriter, r *http.Request) {
	var inputs []communication.Input
	var res Result

	err := json.NewDecoder(r.Body).Decode(&inputs)
	if err != nil {
		log.Debugf("Could not decode provided json : %+v", err)
		sendDefaultValue(w, CodeCouldNotDecodeJSON)
		return
	}

	log.Debugf("JSON input : %+v", inputs)
	defer r.Body.Close()

	/*
	* TODO
	* Implements load balancing betweens node
	* Remove duplications
	* 	- maybe each module should look in the database and check if it has been already done
	* 	- Scan expiration ? re-runable script ? only re run if not in the same area / ip range ?
	 */

	for _, module := range api.Session.ModulesEnabled {
		moduleName := strings.ToLower(module.Name())
		// send as much command as inputs
		for _, input := range inputs {
			cmd := communication.Command{Name: moduleName, Options: input}
			log.Debugf("RunModule for cmd : %+v", cmd)

			err = api.Server.SendCmd(cmd)

			// exit at first error.
			if err != nil {
				sendDefaultValue(w, CodeServerError)
				return
			}
		}
	}

	res = CodeToResult[CodeOK]
	res.Message = "Command sent"
	w.WriteHeader(res.HTTPCode)
	json.NewEncoder(w).Encode(res)
}

/*
RunModule return this template after starting the modules
  "data": {
    nodes: [
      "1.2.3.4",
      "4.3.2.1"
    ]
  }
*/
func (api *API) RunModule(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("RunModule api.db : %+v", api.db)
	var inputs []communication.Input
	var res Result

	vars := mux.Vars(r)
	module := vars["module"]

	err := json.NewDecoder(r.Body).Decode(&inputs)

	if err != nil {
		log.Debugf("Could not decode provided json : %+v", err)
		sendDefaultValue(w, CodeCouldNotDecodeJSON)
		return
	}
	defer r.Body.Close()

	for _, input := range inputs {
		cmd := communication.Command{Name: module, Options: input}
		log.Debugf("RunModule for cmd : %+v", cmd)
		err = api.Server.SendCmd(cmd)
		if err != nil {
			//TODO
			sendDefaultValue(w, CodeNotImplementedYet)
			return
		}
	}

	res = CodeToResult[CodeOK]
	res.Message = "Command sent"
	w.WriteHeader(res.HTTPCode)
	json.NewEncoder(w).Encode(res)
}

//GetIndex returns a link to the documentation on the root path
func (api *API) GetIndex(w http.ResponseWriter, r *http.Request) {

	info := Info{Port: api.Session.Config.API.Port, Versions: Version}
	d := Metadata{Info: info}

	res := CodeToResult[CodeOK]
	res.Data = d
	res.Message = "Documentation available at https://github.com/netm4ul/netm4ul"
	w.WriteHeader(res.HTTPCode)
	json.NewEncoder(w).Encode(res)
}

// GetRoutesByIP returns all the routes informations
func (api *API) GetRoutesByIP(w http.ResponseWriter, r *http.Request) {
	//TODO
	sendDefaultValue(w, CodeNotImplementedYet)
}
