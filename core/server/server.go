package server

import (
	"bufio"
	"encoding/gob"
	"errors"
	"io"
	"net"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/netm4ul/netm4ul/modules"

	"crypto/tls"

	"github.com/netm4ul/netm4ul/core/communication"
	"github.com/netm4ul/netm4ul/core/config"
	"github.com/netm4ul/netm4ul/core/database"
	"github.com/netm4ul/netm4ul/core/database/models"
	"github.com/netm4ul/netm4ul/core/loadbalancing"
	"github.com/netm4ul/netm4ul/core/session"
)

type Server struct {
	//Session represent the server side's session. Hold all the modules
	Session *session.Session
	Db      models.Database
	Algo    loadbalancing.Algorithm
}

// CreateServer : Initialise the infinite server loop on the master node
func CreateServer(s *session.Session) *Server {
	server := Server{Session: s}
	server.Db = database.NewDatabase(&server.Session.Config)

	server.Session.Nodes = make([]communication.Node, 0)
	server.Session.Config.Modules = make(map[string]config.Module)
	algo, err := loadbalancing.NewAlgo()

	if err != nil {
		log.Fatal(err)
	}

	server.Algo = algo

	return &server
}

// Listen : create the TCP server
func (server *Server) Listen() {

	ipport := server.Session.GetServerIPPort()

	var err error
	var l net.Listener

	if server.Session.Config.TLSParams.UseTLS {
		l, err = tls.Listen("tcp", ipport, server.Session.Config.TLSParams.TLSConfig)
	} else {
		l, err = net.Listen("tcp", ipport)
	}

	if err != nil {
		log.Fatalf("Error listening : %s", err.Error())
	}
	// Close the listener when the application closes.
	defer l.Close()

	log.Infof("Listenning on : %s", ipport)

	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			log.Fatalf("Error accepting : %s", err.Error())
		}

		// Handle connections in a new goroutine. (multi-client)
		go server.handleRequest(conn)
	}
}

func (server *Server) handleRequest(conn net.Conn) {

	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	server.handleHello(conn, rw)

	stop := false
	for !stop {
		stop = server.handleData(conn, rw)
	}
}

// Recv basic info for the node at connection time.
func (server *Server) handleHello(conn net.Conn, rw *bufio.ReadWriter) {

	var node communication.Node

	err := gob.NewDecoder(rw).Decode(&node)
	if err != nil {
		log.Errorf("Cannot read hello data : %s", err.Error())
		return
	}
	ip := strings.Split(conn.RemoteAddr().String(), ":")[0]

	node.Conn = conn
	node.IP = ip

	// check if the node is known and create or update it.
	found := false
	var i int
	var n communication.Node
	for i, n = range server.Session.Nodes {
		if n.ID == node.ID {
			found = true
		}
	}

	if found {
		log.Infoln("Node known. Updating")
		server.Session.Nodes[i] = node
	} else {
		server.Session.Nodes = append(server.Session.Nodes, node)
		log.Infoln("Unknown node. Creating")
	}
}

//SendCmdByName is a wrapper to the SendCommand function.
func (server *Server) SendCmdByName(name string, options []string) {
	//TODO get the Command by module name & setup options and requirements

	// cmd := Command{
	// 	Name: name,
	// 	Options: options,
	// 	Requirements: GetRequirementFromCommandName(name)
	// }
	// SendCmd(cmd)
}

//SendCmd sends one commands with its options to selected clients
func (server *Server) SendCmd(command communication.Command) error {

	server.Algo.SetNodes(server.Session.Nodes)
	nodes, err := server.getNextNodes(command)

	log.Debugf("Command (%s) will be executed on %d node(s), Total nodes : %d [ %+v ]",
		command.Name,
		len(nodes),
		len(server.Session.Nodes),
		server.Session.Nodes,
	)

	if err != nil {
		return errors.New("Could not get nodes :" + err.Error())
	}

	// Send to all nodes following the requirements
	for _, node := range nodes {

		rw := bufio.NewReadWriter(bufio.NewReader(node.Conn), bufio.NewWriter(node.Conn))
		err := gob.NewEncoder(rw).Encode(command)

		if err != nil {
			return errors.New("Could not send command :" + err.Error())
		}

		err = rw.Flush()
		if err != nil {
			return errors.New("Could not send command :" + err.Error())
		}
		log.Infof("Sent command \"%s\" to %s", command.Name, node.Conn.RemoteAddr().String())

		node.IsAvailable = false
	}

	return nil
}

//getNextNodes return a list of net.Conn available. They must follows the requirements.
//The next command will be sent on all of these
func (server *Server) getNextNodes(cmd communication.Command) ([]communication.Node, error) {
	// TODO : Requirements for each modules and load balance
	nodes := server.Algo.NextExecutionNodes(cmd)
	log.Debug("Selected nodes : ", nodes)

	return nodes, nil
}

// handleData decode and route all data after the "hello". It listens forever until connection closed.
func (server *Server) handleData(conn net.Conn, rw *bufio.ReadWriter) bool {
	var data modules.Result

	err := gob.NewDecoder(rw).Decode(&data)

	// handle connection closed (client shutdown)
	if err == io.EOF {
		log.Errorf("Connection closed : %s", err.Error())
		// stop all handleData for this conn
		return true
	}

	// handle other error
	if err != nil {
		log.Errorf("Error while decoding data : %s", err.Error())
		return false
	}
	log.Debugf("%+v", data)

	module, ok := server.Session.Modules[strings.ToLower(data.Module)]
	if !ok {
		log.Errorf("Unknown module : %s %+v %+v", data.Module, module, ok)
	}

	ip := strings.Split(conn.RemoteAddr().String(), ":")[0]

	found := false
	var node communication.Node
	var i int
	for i, node = range server.Session.Nodes {
		if node.IP == ip {
			found = true
		}
	}

	if !found {
		log.Error("Coulnd't get this node, unknown ip :", ip)
		return false
	}

	server.Db.Connect(&server.Session.Config)
	//update it every time
	p := models.Project{Name: server.Session.Nodes[i].Project}
	server.Db.CreateOrUpdateProject(p)

	err = module.WriteDb(data, server.Db, p.Name)

	if err != nil {
		log.Errorf("Database error : %+v", err)
	}
	log.Infof("Saved database info, module : %s", data.Module)
	return false
}
