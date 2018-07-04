package api

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

//UnauthentificatedWhiteList contains all of the allowed ressources without MAC, unprefixed
var UnauthentificatedWhiteList = []string{
	"/",
	"/users/create",
	"/users/login",
}

func (api *API) isUnderMAC(r *http.Request) bool {
	//search in the whitelist
	for _, whitelisted := range UnauthentificatedWhiteList {
		if api.Prefix+whitelisted == r.URL.String() {
			return false
		}
	}

	//route not in the whitelist !
	return true
}

func (api *API) jsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debugf("Request URL : %s", r.RequestURI)
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func (api *API) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		//bypass authentification for whitelisted api calls
		if !api.isUnderMAC(r) {
			next.ServeHTTP(w, r)
		}

		token := r.Header.Get("X-Session-Token")
		log.Debugf("Token : %s", token)
		user, err := api.db.GetUserByToken(token)

		if err != nil {
			log.Errorf("err : %+v", err)
			sendDefaultValue(w, CodeForbidden)
			//ensure we don't write anymore to w
			return
		}

		if user.ID != "" {
			log.Debugf("Authenticated user [%s]: %s\n", user.ID, user.Name)
			next.ServeHTTP(w, r)
		} else {
			sendDefaultValue(w, CodeForbidden)
			// http.Error(w, AccessForbiddenResponse, http.StatusForbidden)
		}
	})
}
