package config

import (
	"testing"
)

func init() {
	LoadConfig("../../netm4ul.conf")
}

func TestParseProject(t *testing.T) {
	if Config.Project.Name == "" {
		t.Error("Expected project name, got nothing")
	}
}

func TestParseAPI(t *testing.T) {
	var port uint16

	user := Config.API.User
	password := Config.API.Password

	port = Config.API.Port

	if user == "" {
		t.Error("Expected user, got empty string")
	}
	if port == 0 {
		t.Error("Expected port, got 0")
	}
	if password == "" {
		t.Error("Expected password, got empty string")
	}
}

func TestParseServer(t *testing.T) {

	user := Config.Server.User
	password := Config.Server.Password
	ip := Config.Server.IP
	port := Config.Server.Port

	if user == "" {
		t.Error("Expected user, got ", user)
	}
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

	usetls := Config.TLSParams.UseTLS
	caCert := Config.TLSParams.CaCert
	serverCert := Config.TLSParams.ServerCert
	serverPrivateKey := Config.TLSParams.ServerPrivateKey
	clientCert := Config.TLSParams.ClientCert
	clientPrivateKey := Config.TLSParams.ClientPrivateKey

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

	database := Config.Database.Database
	user := Config.Database.User
	password := Config.Database.Password
	ip := Config.Database.IP
	port := Config.Database.Port

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
	moduleCount := len(Config.Modules)
	if moduleCount == 0 {
		t.Error("Expected at least one module, got 0")
	}
	if !Config.Modules["traceroute"].Enabled {
		t.Error("Expected traceroute to be enabled", Config.Modules["traceroute"].Enabled)
	}
}
