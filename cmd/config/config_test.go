package config

import (
	"net"
	"testing"
)

func TestParseServer(t *testing.T) {
	var user string
	var password string
	var port uint16
	var ip net.IP

	user = Config.Server.User
	password = Config.Server.Password
	ip = Config.Server.IP
	port = Config.Server.Port

	if user != "user" {
		t.Error("Expected 'user', got ", user)
	}
	if password != "password" {
		t.Error("Expected 'password', got ", password)
	}
	if !ip.Equal(net.ParseIP("127.0.0.1")) {
		t.Error("Expected net.IP('127.0.0.1'), got ", ip)
	}
	if port != 5672 {
		t.Error("Expected 5672, got ", port)
	}
}

func TestParseAPI(t *testing.T) {
	var port uint16
	var user string
	var password string

	user = Config.API.User
	port = Config.API.Port
	password = Config.API.Password

	if user != "toto" {
		t.Error("Expected 'toto', got ", user)
	}
	if port != 8080 {
		t.Error("Expected 8080, got ", port)
	}
	if password != "P@ssW0rd!" {
		t.Error("Expected 'P@ssW0rd!', got ", password)
	}
}

func TestModules(t *testing.T) {
	//	var enabled bool
	moduleCount := len(Config.Modules)
	if moduleCount != 3 {
		t.Error("Expected 3 modules, got", moduleCount)
	}
	if !Config.Modules["shodan"].Enabled {
		t.Error("Expected shodan to be enabled", Config.Modules["shodan"].Enabled)
	}
}
