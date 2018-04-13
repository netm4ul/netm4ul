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
	"crypto/tls"
)

const (
	maxRetry = 3
)

// CreateServer : Initialise the infinite server loop on the master node
func CreateServer(ipport string, conf *config.ConfigToml) {
	server.ConfigServer = conf
	server.TLSListen(ipport, conf.TLSParams)
}

// CreateAPI : Initialise the infinite server loop on the master node
func CreateAPI(ipport string, conf *config.ConfigToml) {
	// api.ConfigServer = conf
	api.Start(ipport, conf)
}

// CreateClient : Connect the node to the master server
func CreateClient(ipport string, conf *config.ConfigToml) {
	client.InitModule()


	log.Println(colors.Green("Modules enabled :"), client.ListModuleEnabled)
	var err error
	var conn *tls.Conn

	for tries := 0; tries < maxRetry; tries++ {
		conn, err = client.Connect(ipport, conf.TLSParams)
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
