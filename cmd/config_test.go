package cmd

import(
	"testing"
	"net"
)

func TestParseMQ(t *testing.T){
	var user string
	var password string
	var port uint16
	var ip net.IP

	user = Config.MQ.User
	password = Config.MQ.Password
	ip = Config.MQ.Ip
	port = Config.MQ.Port

	if user != "user"{
		t.Error("Expected 'user', got ", user)
	}
	if password != "password"{
		t.Error("Expected 'password', got ", password)
	}
	if !ip.Equal(net.ParseIP("1.1.1.1")){
		t.Error("Expected net.IP('1.1.1.1'), got ", ip)
	}
	if port != 5672{
		t.Error("Expected 5672, got ", port)
	}
}

func TestParseApi(t *testing.T){
	var port uint16
	var user string
	var password string

	user = Config.Api.User
	port = Config.Api.Port
	password = Config.Api.Password

	if user != "toto"{
		t.Error("Expected 'toto', got ", user)
	}
	if port != 8080{
		t.Error("Expected 8080, got ", port)
	}
	if password != "P@ssW0rd!"{
		t.Error("Expected 'P@ssW0rd!', got ", password)
	}
}
