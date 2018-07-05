package config

import (
	"testing"
)

var config ConfigToml

func init() {
	var err error
	config, err = LoadConfig("../../netm4ul.conf")
	if err != nil {
		panic("Could not load config for testing !")
	}
}

func TestParseProject(t *testing.T) {
	if config.Project.Name == "" {
		t.Error("Expected project name, got nothing")
	}
}

func TestParseAPI(t *testing.T) {
	var port uint16

	user := config.API.User
	token := config.API.Token

	port = config.API.Port

	if user == "" {
		t.Error("Expected user, got empty string")
	}
	if port == 0 {
		t.Error("Expected port, got 0")
	}
	if token == "" {
		t.Log("Expected token, got empty string")
	}
}

func TestParseServer(t *testing.T) {

	password := config.Server.Password
	ip := config.Server.IP
	port := config.Server.Port

	if password == "" {
		t.Error("Expected password, got ", password)
	}
	if ip == "" {
		t.Error("Expected ip or domain, got ", ip)
	}
	if port == 0 {
		t.Error("Expected port number, got ", port)
	}
}

func TestParseTLS(t *testing.T) {

	usetls := config.TLSParams.UseTLS
	caCert := config.TLSParams.CaCert
	serverCert := config.TLSParams.ServerCert
	serverPrivateKey := config.TLSParams.ServerPrivateKey
	clientCert := config.TLSParams.ClientCert
	clientPrivateKey := config.TLSParams.ClientPrivateKey

	// Test only if the config is using TLS.
	// Don't want false fail, but might reconsider
	if usetls {
		if caCert == "" {
			t.Error("Expected Root CA certificates path, got empty path")
		}
		if serverCert == "" {
			t.Error("Expected Server certificates path, got empty path")
		}
		if serverPrivateKey == "" {
			t.Error("Expected Server private key path, got empty path")
		}
		if clientCert == "" {
			t.Error("Expected Client certificate path, got empty path")
		}
		if clientPrivateKey == "" {
			t.Error("Expected Client private key path, got empty path")
		}
	}
}

func TestParseDatabase(t *testing.T) {

	database := config.Database.Database
	user := config.Database.User
	password := config.Database.Password
	ip := config.Database.IP
	port := config.Database.Port

	if database == "" {
		t.Error("Expected database name, got empty string")
	}
	if user == "" {
		t.Error("Expected user, got empty string")
	}
	if password == "" {
		t.Error("Expected password, got empty string")
	}
	if ip == "" {
		t.Error("Expected ip or domain, got empty string")
	}
	if port == 0 {
		t.Error("Expected port number, got 0")
	}
}

func TestModules(t *testing.T) {
	//	var enabled bool
	moduleCount := len(config.Modules)
	if moduleCount == 0 {
		t.Error("Expected at least one module, got 0")
	}
	if !config.Modules["traceroute"].Enabled {
		t.Error("Expected traceroute to be enabled", config.Modules["traceroute"].Enabled)
	}
}
