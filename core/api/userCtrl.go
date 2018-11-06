package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/netm4ul/netm4ul/core/database/models"
	"github.com/netm4ul/netm4ul/core/security"
	log "github.com/sirupsen/logrus"
)

type simplifiedUser struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

func (api *API) isCorrectPassword(username string, plain string) (bool, models.User, error) {

	user, err := api.db.GetUser(username)
	if err != nil {
		return false, models.User{}, err
	}

	if security.ComparePassword(user.Password, plain) {
		return true, user, nil
	}

	return false, models.User{}, nil

}

//CreateUser handle the requests for creating users.
func (api *API) CreateUser(w http.ResponseWriter, r *http.Request) {
	var res Result
	var user simplifiedUser

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&user)

	if err != nil {
		sendDefaultValue(w, CodeCouldNotDecodeJSON)
		return
	}
	defer r.Body.Close()

	userDB, err := api.db.GetUser(user.Name)
	if err != nil {
		sendDefaultValue(w, CodeServerError)
		return
	}

	if userDB.Name != "" {
		res = CodeToResult[CodeInvalidInput]
		res.Message = "User already exist"
		w.WriteHeader(CodeToResult[CodeInvalidInput].HTTPCode)
		json.NewEncoder(w).Encode(res)
		return
	}

	pass, err := security.HashAndSalt([]byte(user.Password))
	if err != nil {
		sendDefaultValue(w, CodeServerError)
		return
	}
	newUser := models.User{
		Name:     user.Name,
		Password: pass,
		Token:    security.GenerateNewToken(),
	}

	err = api.db.CreateOrUpdateUser(newUser)
	if err != nil {
		sendDefaultValue(w, CodeServerError)
		return
	}

	res = CodeToResult[CodeOK]
	res.Data = struct{ Token string }{Token: newUser.Token}
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		sendDefaultValue(w, CodeServerError)
		return
	}
}

//GetUser will return basic information on the user
func (api *API) GetUser(w http.ResponseWriter, r *http.Request) {
	var res Result

	vars := mux.Vars(r)
	user, err := api.db.GetUser(vars["name"])

	if err != nil {
		log.Warnf("User not found %s", vars["name"])
		res = CodeToResult[CodeNotFound]
		res.Message = "User not found"
		w.WriteHeader(CodeToResult[CodeNotFound].HTTPCode)
		json.NewEncoder(w).Encode(res)
		return
	}

	//Do not report user private info
	user.Password = ""
	user.Token = ""
	res = CodeToResult[CodeOK]
	res.Data = user

	json.NewEncoder(w).Encode(res)
}

/*
UserLogin will return a new token if the provided user match an existing one
You must send the following json with correct user/password :
```
{
	"name":"someusername",
	"password":"somepassword"
}
```
*/
func (api *API) UserLogin(w http.ResponseWriter, r *http.Request) {

	var res Result
	var user simplifiedUser

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&user)

	if err != nil {
		sendDefaultValue(w, CodeCouldNotDecodeJSON)
		return
	}
	defer r.Body.Close()

	ok, loggedUser, err := api.isCorrectPassword(user.Name, user.Password)
	if err != nil {
		sendDefaultValue(w, CodeServerError)
		return
	}

	if !ok {
		sendDefaultValue(w, CodeForbidden)
		return
	}

	res = CodeToResult[CodeOK]
	res.Data = loggedUser.Token
	json.NewEncoder(w).Encode(res)
}

//UserLogout disable the provided token
func (api *API) UserLogout(w http.ResponseWriter, r *http.Request) {

	var res Result

	token := r.Header.Get("X-Session-Token")
	user, err := api.db.GetUserByToken(token)

	if err != nil {
		log.Errorf("Error : %s", err)
		sendDefaultValue(w, CodeServerError)
		return
	}

	user.Token = ""
	err = api.db.CreateOrUpdateUser(user)

	if err != nil {
		log.Errorf("Error : %s", err)
		sendDefaultValue(w, CodeServerError)
		return
	}

	res = CodeToResult[CodeOK]
	res.Data = ""
	res.Message = "Logged out"
	json.NewEncoder(w).Encode(res)
}
