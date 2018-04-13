package core

import (
	"io"
	"log"
	"net"
	"time"

	"github.com/netm4ul/netm4ul/cmd/colors"
	"github.com/netm4ul/netm4ul/core/api"
	"github.com/netm4ul/netm4ul/core/client"
	"github.com/netm4ul/netm4ul/core/config"
	"github.com/netm4ul/netm4ul/core/server"
	"github.com/netm4ul/netm4ul/core/session"
)

const (
	maxRetry = 3
)

// CreateServer : Initialise the infinite server loop on the master node
func CreateServer(conf config.ConfigToml) {
	s := session.NewSession(conf)
	server.InitServer(s)
	server.Listen()
}

// CreateAPI : Initialise the infinite server loop on the master node
func CreateAPI(conf config.ConfigToml) {
	s := session.NewSession(conf)
	api.InitApi(s)
	api.Start()
}

// CreateClient : Connect the node to the master server
func CreateClient(conf config.ConfigToml) {

	s := session.NewSession(conf)
	client.Create(s)

	log.Println(colors.Green("Modules enabled :"), client.ListModuleEnabled)
	var err error
	var conn *net.TCPConn

	for tries := 0; tries < maxRetry; tries++ {
		ipport := s.GetServerIPPort()
		conn, err = client.Connect(ipport)
		if err != nil {
			log.Println(colors.Red("Could not connect :"), err)
			log.Printf(colors.Red("Retry count : %d, Max retry : %d"), tries, maxRetry)
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}

	if err != nil {
		log.Fatal(err)
	}

	err = client.SendHello(conn)
	if err != nil {
		log.Fatal(err)
	}

	// Recieve data
	go func() {
		for {
			cmd, err := client.Recv(conn)

			// kill on socket closed.
			if err == io.EOF {
				log.Fatalf(colors.Red("Connection closed : %s"), err.Error())
			}

			if err != nil {
				log.Println(colors.Red(err.Error()))
				continue
			}

			// must exist, if it doesn't, this line shouldn't be executed (checks above)
			module := client.SessionClient.Modules[cmd.Name]

			//TODO
			// send data back to the server
			data, err := client.Execute(module, cmd)
			client.SendResult(conn, data)
		}
	}()
}
