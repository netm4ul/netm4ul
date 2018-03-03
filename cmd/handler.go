package cmd

import (
	"bufio"
	"log"
	"time"

	"github.com/netm4ul/netm4ul/cmd/client"
	"github.com/netm4ul/netm4ul/cmd/config"
	"github.com/netm4ul/netm4ul/cmd/server"
)

const (
	maxRetry = 3
)

// CreateServer : Initialise the infinite server loop on the master node
func CreateServer(ipport string, conf *config.ConfigToml) {
	server.ConfigServer = conf
	server.Listen(ipport)
}

// CreateClient : Connect the node to the master server
func CreateClient(ipport string, conf *config.ConfigToml) {
	client.InitModule()
	var err error
	var rw *bufio.ReadWriter

	for tries := 0; tries < maxRetry; tries++ {
		rw, err = client.Connect(ipport)
		if err != nil {
			log.Println("Could not connect : ", err)
			log.Println("Retry count : ", tries, "Max retry : ", maxRetry)
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}

	if err != nil {
		log.Fatal(err)
	}

	err = client.SendHello(rw)
	if err != nil {
		log.Fatal(err)
	}
	// TODO : Client.Recv(cmd) & Client.Send(data)
}
