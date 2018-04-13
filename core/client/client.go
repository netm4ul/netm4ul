package client

import (
	"bufio"
	"encoding/gob"
	"io"
	"log"
	"github.com/netm4ul/netm4ul/cmd/colors"
	"github.com/netm4ul/netm4ul/core/config"
	"github.com/netm4ul/netm4ul/core/server"
	"github.com/netm4ul/netm4ul/core/session"
	"github.com/netm4ul/netm4ul/modules"
	"github.com/pkg/errors"
	"crypto/tls"
)

const (
	Version = "0.1"
)

var (
	SessionClient *session.Session
	// ListModule : global list of modules
	ListModule []string
	// ListModuleEnabled : global list of enabled modules
	ListModuleEnabled []string
)

// Connect : Setup the connection to the master node
func Connect(ipport string, config tls.Config) (*tls.Conn, error) {
	conn, err := tls.Dial("tcp", ipport, &config)
	// conn.SetKeepAlive(true) // This could potentially be buggy : tls.Conn does not offer a SetKeepAllive equivalent

	if err != nil {
		return nil, errors.Wrap(err, "Dialing "+ipport+" failed")
	}

	return conn, nil
}

// InitModule : Update ListModule & ListModuleEnabled variable
func InitModule() {
	SessionClient = session.NewSession()
	if config.Config.Verbose {
		log.Printf(colors.Yellow("Session client : %+v"), SessionClient)
	}
	for m := range config.Config.Modules {
		ListModule = append(ListModule, m)
		if config.Config.Modules[m].Enabled {
			ListModuleEnabled = append(ListModuleEnabled, m)
		}
	}
}

// SendHello : Send node info (modules list, project name,...)
func SendHello(conn *tls.Conn) error {
	var err error
	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	enc := gob.NewEncoder(rw)

	module := ListModuleEnabled
	node := config.Node{Modules: module, Project: "FirstProject"}

	if config.Config.Verbose {
		log.Printf(colors.Yellow("Node : %+v"), node)
	}

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

// Recv read the incomming data from the server. The server use the server.Command struct.
func Recv(conn *tls.Conn) (server.Command, error) {
	var cmd server.Command

	if config.Config.Verbose {
		log.Println(colors.Yellow("Waiting for incoming data"))
	}

	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	dec := gob.NewDecoder(rw)
	err := dec.Decode(&cmd)

	// handle connection closed (server shutdown for example)
	if err == io.EOF {
		return server.Command{}, err
	}

	if err != nil {
		return server.Command{}, errors.New("Could not decode received message : " + err.Error())
	}

	if config.Config.Verbose {
		log.Printf(colors.Yellow("Received command %+v"), cmd)
	}
	_, ok := SessionClient.Modules[cmd.Name]

	if !ok {
		return server.Command{}, errors.New("Unsupported (or unknown) command : " + cmd.Name)
	}

	return cmd, nil
}

// Execute runs modules with the right options and handle errors.
func Execute(module modules.Module, cmd server.Command) (modules.Result, error) {

	if config.Config.Verbose {
		log.Printf("Executing module : \n\t %s, version %s by %s\n\t", module.Name(), module.Version(), module.Author())
	}
	//TODO
	res, err := module.Run(cmd.Options)
	return res, err
}

// SendResult sends the data back to the server. It will then be handled by each module.WriteDb to be saved
func SendResult(conn *tls.Conn, res modules.Result) error {
	var err error
	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	enc := gob.NewEncoder(rw)
	err = enc.Encode(res)

	if err != nil {
		log.Println(colors.Red("Error :"), err)
		return err
	}

	err = rw.Flush()
	if err != nil {
		log.Println(colors.Red("Error :"), err)
		return err
	}

	return nil

}
