package api_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/netm4ul/netm4ul/core/config"
	"github.com/netm4ul/netm4ul/core/server"
	"github.com/netm4ul/netm4ul/core/session"

	"github.com/netm4ul/netm4ul/core/api"
)

var (
	testserver *httptest.Server
	reader     io.Reader
	a          *api.API
)

func init() {

	//TODO set a default config !
	conf := config.ConfigToml{}

	sess := session.NewSession(conf)
	serv := server.CreateServer(sess)

	a = api.CreateAPI(sess, serv)

}

func TestHTTPIndex(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/v1", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(a.GetIndex)

	handler.ServeHTTP(rr, req)

	var jsonres api.Result

	err = json.NewDecoder(rr.Body).Decode(&jsonres)

	if err != nil {
		t.Errorf("Could not decode json ! : %+v", err)
	}

	if rr.Code != http.StatusOK {
		t.Errorf("Bad HTTP Code, got %d instead of %d", rr.Code, http.StatusOK)
	}

	if jsonres.Code != api.CodeOK {
		t.Errorf("Bad response code, got %d instead of %d", jsonres.Code, api.CodeOK)
	}

	expected := "Documentation available at https://github.com/netm4ul/netm4ul"

	if jsonres.Message != expected {
		t.Errorf("Expected %s, got %s", expected, jsonres.Message)
	}
}
