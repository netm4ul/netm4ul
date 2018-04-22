package client

import (
	"bufio"
	"crypto/tls"
	"encoding/gob"
	"io"
	"net"
	"time"

	"github.com/netm4ul/netm4ul/core/config"
	"github.com/netm4ul/netm4ul/core/server"
	"github.com/netm4ul/netm4ul/core/session"
	"github.com/netm4ul/netm4ul/modules"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	Version  = "0.1"
	maxRetry = 3
)

var (
	//SessionClient : !! TO FIX : global var for state vs args of each func ?
	SessionClient *session.Session
	// ListModule : global list of modules
	ListModule []string
	// ListModuleEnabled : global list of enabled modules
	ListModuleEnabled []string
)

// CreateClient : Connect the node to the master server
func CreateClient(s *session.Session) {

	var err error
	SessionClient = s

	InitModule(s)

	log.Info("Modules enabled :", ListModuleEnabled)

	for tries := 0; tries < maxRetry; tries++ {
		err = Connect(s)

		// no error, exit retry loop
		if err == nil {
			break
		}

		log.Errorf("Could not connect : %+v", err)
		log.Errorf("Retry count : %d, Max retry : %d", tries, maxRetry)
		time.Sleep(10 * time.Second)
	}

	if err != nil {
		log.Fatal(err)
	}

	err = SendHello(s)
	if err != nil {
		log.Fatal(err)
	}

	// Recieve data
	go handleData(s)
}

func handleData(s *session.Session) {

	for {
		cmd, err := Recv(s)

		// kill on socket closed.
		if err == io.EOF {
			log.Fatalf("Connection closed : %s", err.Error())
		}

		if err != nil {
			log.Error(err.Error())
			continue
		}

		// must exist, if it doesn't, this line shouldn't be executed (checks above)
		module := s.Modules[cmd.Name]

		//TODO
		// send data back to the server
		data, err := Execute(s, module, cmd)
		SendResult(s, data)
	}
}

// Connect : Setup the connection to the master node
func Connect(s *session.Session) error {

	var err error
	ipport := s.GetServerIPPort()

	if s.Config.TLSParams.UseTLS {
		s.Connector.TLSConn, err = tls.Dial("tcp", ipport, s.Config.TLSParams.TLSConfig)
		if err != nil {
			return errors.Wrap(err, "Dialing "+ipport+" failed")
		}
	} else {
		s.Connector.Conn, err = net.Dial("tcp", ipport)
		if err != nil {
			return errors.Wrap(err, "Dialing "+ipport+" failed")
		}
	}

	return nil

}

// InitModule : Update ListModule & ListModuleEnabled variable
func InitModule(s *session.Session) {

	log.Debugf("Session client : %+v", s)
	for m := range s.Config.Modules {
		ListModule = append(ListModule, m)
		if s.Config.Modules[m].Enabled {
			ListModuleEnabled = append(ListModuleEnabled, m)
		}
	}
}

// SendHello : Send node info (modules list, project name,...)
func SendHello(s *session.Session) error {
	var rw *bufio.ReadWriter

	if s.Connector.TLSConn == nil {
		rw = bufio.NewReadWriter(bufio.NewReader(s.Connector.Conn), bufio.NewWriter(s.Connector.Conn))
	} else {
		rw = bufio.NewReadWriter(bufio.NewReader(s.Connector.TLSConn), bufio.NewWriter(s.Connector.TLSConn))
	}

	enc := gob.NewEncoder(rw)

	node := config.Node{Modules: ListModuleEnabled, Project: s.Config.Project.Name}

	log.Debugf("Node : %+v", node)

	err := enc.Encode(node)
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
func Recv(s *session.Session) (server.Command, error) {
	var cmd server.Command

	log.Debugf("Waiting for incomming data")

	var rw *bufio.ReadWriter

	if s.Connector.TLSConn == nil {
		rw = bufio.NewReadWriter(bufio.NewReader(s.Connector.Conn), bufio.NewWriter(s.Connector.Conn))
	} else {
		rw = bufio.NewReadWriter(bufio.NewReader(s.Connector.TLSConn), bufio.NewWriter(s.Connector.TLSConn))
	}

	dec := gob.NewDecoder(rw)
	err := dec.Decode(&cmd)

	// handle connection closed (server shutdown for example)
	if err == io.EOF {
		return server.Command{}, err
	}

	if err != nil {
		return server.Command{}, errors.New("Could not decode received message : " + err.Error())
	}

	log.Debugf("Recieved command %+v", cmd)
	_, ok := s.Modules[cmd.Name]

	if !ok {
		return server.Command{}, errors.New("Unsupported (or unknown) command : " + cmd.Name)
	}

	return cmd, nil
}

// Execute runs modules with the right options and handle errors.
func Execute(s *session.Session, module modules.Module, cmd server.Command) (modules.Result, error) {

	log.Debugf("Executing module : \n\t %s, version %s by %s\n\t", module.Name(), module.Version(), module.Author())
	//TODO
	res, err := module.Run(cmd.Options)
	return res, err
}

// SendResult sends the data back to the server. It will then be handled by each module.WriteDb to be saved
func SendResult(s *session.Session, res modules.Result) error {
	var rw *bufio.ReadWriter

	if s.Connector.TLSConn == nil {
		rw = bufio.NewReadWriter(bufio.NewReader(s.Connector.Conn), bufio.NewWriter(s.Connector.Conn))
	} else {
		rw = bufio.NewReadWriter(bufio.NewReader(s.Connector.TLSConn), bufio.NewWriter(s.Connector.TLSConn))
	}

	enc := gob.NewEncoder(rw)
	err := enc.Encode(res)

	if err != nil {
		log.Errorf("Error : %+v", err)
		return err
	}

	err = rw.Flush()
	if err != nil {
		log.Errorf("Error : %+v", err)
		return err
	}

	return nil
}
