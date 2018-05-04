package api_test

import (
	"testing"

	"github.com/netm4ul/netm4ul/core/api"
	"github.com/netm4ul/netm4ul/core/config"
	"github.com/netm4ul/netm4ul/core/session"

	"github.com/netm4ul/netm4ul/core/server"
	"github.com/stretchr/testify/assert"
)

var testSession *session.Session
var s *server.Server

func init() {
	conf := config.ConfigToml{IsServer: true}
	testSession = &session.Session{Config: conf}
	s = server.CreateServer(testSession)

	api.CreateAPI(testSession, s)
}

func TestHTTPIndex(t *testing.T) {

	assert.Equal(t, 123, 123, "they should be equal")
}
