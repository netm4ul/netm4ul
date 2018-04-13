package config

import (
	"testing"
)

func init() {
	LoadConfig("../../netm4ul.conf")
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

func TestParseAPI(t *testing.T) {
	var port uint16

	user := Config.API.User
	password := Config.API.Password

	port = Config.API.Port

	if user == "" {
		t.Error("Expected user, got ", user)
	}
	if port == 0 {
		t.Error("Expected port, got ", port)
	}
	if password == "" {
		t.Error("Expected password, got ", password)
	}
}

func TestModules(t *testing.T) {
	//	var enabled bool
	moduleCount := len(Config.Modules)
	if moduleCount == 0 {
		t.Error("Expected at least one module, got", moduleCount)
	}
	if !Config.Modules["traceroute"].Enabled {
		t.Error("Expected traceroute to be enabled", Config.Modules["traceroute"].Enabled)
	}
}
