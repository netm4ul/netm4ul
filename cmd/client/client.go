package client

import (
	"bufio"
	"encoding/gob"
	"log"
	"net"

	"github.com/netm4ul/netm4ul/cmd/config"
	"github.com/netm4ul/netm4ul/cmd/server"
	"github.com/pkg/errors"
)

var (
	// ListModule : global list of modules
	ListModule []string
	// ListModuleEnabled : global list of enabled modules
	ListModuleEnabled []string
)

// Connect : Setup the connection to the master node
func Connect(ipport string) (*bufio.ReadWriter, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", ipport)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	conn.SetKeepAlive(true)

	if err != nil {
		return nil, errors.Wrap(err, "Dialing "+ipport+" failed")
	}

	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	return rw, nil
}

// InitModule : Update ListModule & ListModuleEnabled variable
func InitModule() {
	for m := range config.Config.Modules {
		ListModule = append(ListModule, m)
		if config.Config.Modules[m].Enabled {
			ListModuleEnabled = append(ListModuleEnabled, m)
		}
	}
}

// SendHello : Send node info (modules list, project name,...)
func SendHello(rw *bufio.ReadWriter) error {
	var err error

	enc := gob.NewEncoder(rw)

	module := ListModuleEnabled
	node := config.Node{Modules: module, Project: "FirstProject"}

	log.Println(node)

	err = enc.Encode(node)
	if err != nil {
		return err
	}

	err = rw.Flush()
	if err != nil {
		return err
	}

	return nil
}

//Recv read the incomming data from the server. The server use the server.Command struct.
func Recv(rw *bufio.ReadWriter) (server.Command, error) {
	log.Println("Waiting for incomming data")
	var cmd server.Command
	dec := gob.NewDecoder(rw)
	err := dec.Decode(&cmd)

	if err != nil {
		return server.Command{}, errors.New("Could not decode recieved message")
	}
	log.Printf("Recieved command %+v", cmd)

	return server.Command{}, errors.New("Could not decode recieved message")
}
