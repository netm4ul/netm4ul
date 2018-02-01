package cmd

import (
	"log"

	"github.com/netm4ul/netm4ul/cmd/client"
	"github.com/netm4ul/netm4ul/cmd/config"
	"github.com/netm4ul/netm4ul/cmd/server"
)

// CreateServer : Initialise the infinite server loop on the master node
func CreateServer(ipport string, conf *config.ConfigToml) {
	server.ConfigServer = conf
	server.Listen(ipport)
}

// CreateClient : Connect the node to the master server
func CreateClient(ipport string, conf *config.ConfigToml) {
	client.InitModule()

	rw, err := client.Connect(ipport)

	if err != nil {
		log.Fatal(err)
	}

	err = client.SendHello(rw)
	if err != nil {
		log.Fatal(err)
	}
	// TODO : Client.Recv(cmd) & Client.Send(data)
}
