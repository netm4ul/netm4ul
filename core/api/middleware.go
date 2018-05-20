package api

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

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

		token := r.Header.Get("X-Session-Token")
		user, err := api.db.GetUserByToken(token)

		if err != nil {
			http.Error(w, "Forbidden", http.StatusForbidden)
		}

		if user.ID != "" {
			log.Printf("Authenticated user %+v\n", user)
			next.ServeHTTP(w, r)
		} else {
			http.Error(w, "Forbidden", http.StatusForbidden)
		}
	})
}
