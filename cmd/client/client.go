package client

import (
	"bufio"
	"encoding/gob"
	"io"
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
func Connect(ipport string) (*net.TCPConn, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", ipport)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	conn.SetKeepAlive(true)

	if err != nil {
		return nil, errors.Wrap(err, "Dialing "+ipport+" failed")
	}

	return conn, nil
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
func SendHello(conn *net.TCPConn) error {
	var err error
	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

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
func Recv(conn *net.TCPConn) (server.Command, error) {
	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	log.Println("Waiting for incomming data")
	var cmd server.Command
	dec := gob.NewDecoder(rw)
	err := dec.Decode(&cmd)

	// handle connection closed (server shutdown for example)
	if err == io.EOF {
		return server.Command{}, errors.New("Connection closed : " + err.Error())
	}

	if err != nil {
		return server.Command{}, errors.New("Could not decode recieved message : " + err.Error())
	}

	log.Printf("Recieved command %+v", cmd)

	return server.Command{}, errors.New("Could not decode recieved message")
}
