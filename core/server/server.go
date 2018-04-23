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
	mgo "gopkg.in/mgo.v2"

	"crypto/tls"

	"github.com/netm4ul/netm4ul/core/config"
	"github.com/netm4ul/netm4ul/core/database"
	"github.com/netm4ul/netm4ul/core/requirements"
	"github.com/netm4ul/netm4ul/core/session"
)

type Server struct {
	//Nodes represent a map to net.Conn
	Nodes map[string]net.Conn
	//Session represent the server side's session. Hold all the modules
	Session *session.Session
}

//Command represents the communication protocol between clients and the master node
type Command struct {
	Name         string                    `json:"name"`
	Options      []modules.Input           `json:"options"`
	Requirements requirements.Requirements `json:"requirements"`
}

// CreateServer : Initialise the infinite server loop on the master node
func CreateServer(s *session.Session) *Server {
	server := Server{Nodes: make(map[string]net.Conn), Session: s}
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

	database.InitDatabase(&server.Session.Config)
	mgoSession := database.Connect()

	log.Infof("Listenning on : %s", ipport)

	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			log.Fatalf("Error accepting : %s", err.Error())
		}

		mgoSessionClone := mgoSession.Clone()
		// Handle connections in a new goroutine. (multi-client)
		go server.handleRequest(conn, mgoSessionClone)
	}
}

func (server *Server) handleRequest(conn net.Conn, mgoSession *mgo.Session) {

	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	server.handleHello(conn, rw, mgoSession)

	stop := false
	for !stop {
		stop = server.handleData(conn, rw, mgoSession)
	}
}

// Recv basic info for the node at connection time.
func (server *Server) handleHello(conn net.Conn, rw *bufio.ReadWriter, mgoSession *mgo.Session) {

	var node config.Node

	err := gob.NewDecoder(rw).Decode(&node)

	if err != nil {
		log.Errorf("Cannot read hello data : %s", err.Error())
		return
	}

	ip := strings.Split(conn.RemoteAddr().String(), ":")[0]

	if server.Session.Config.Verbose {
		if _, ok := server.Session.Config.Nodes[ip]; ok {
			log.Infoln("Node known. Updating")
		} else {
			log.Infoln("Unknown node. Creating")
		}
	}

	server.Session.Config.Nodes = make(map[string]config.Node)
	server.Session.Config.Modules = make(map[string]config.Module)
	server.Session.Config.Nodes[ip] = node

	server.Nodes[ip] = conn
	database.CreateProject(mgoSession, node.Project)

	p := database.GetProjects(mgoSession)

	log.Debugf("Nodes : %+v", server.Session.Config.Nodes)
	log.Debugf("Projects : %+v", p)

}

func (server *Server) getProjectByNodeIP(ip string) (string, error) {

	var err error

	n, ok := server.Session.Config.Nodes[ip]
	if !ok {
		return "", errors.New("Unknown node ! Could not get project name")
	}

	project := n.Project

	return project, err
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
func (server *Server) SendCmd(command Command) error {

	conns, err := server.getAvailableNodes(command.Requirements)

	log.Debugf("Available node(s) : %d", len(conns))

	if err != nil {
		return errors.New("Could not get nodes :" + err.Error())
	}

	// Send to all nodes following the requirements
	for _, conn := range conns {

		rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
		err := gob.NewEncoder(rw).Encode(command)

		if err != nil {
			return errors.New("Could not send command :" + err.Error())
		}

		err = rw.Flush()
		if err != nil {
			return errors.New("Could not send command :" + err.Error())
		}
		log.Infof("Sent command \"%s\" to %s", command.Name, conn.RemoteAddr().String())
	}

	return nil
}

//getAvailableNodes return a list of net.Conn available. They must follows the requirements.
func (server *Server) getAvailableNodes(req requirements.Requirements) ([]net.Conn, error) {
	// TODO : Requirements for each modules and load balance
	var availables []net.Conn
	for _, conn := range server.Nodes {
		availables = append(availables, conn)
	}
	return availables, nil
}

// handleData decode and route all data after the "hello". It listens forever until connection closed.
func (server *Server) handleData(conn net.Conn, rw *bufio.ReadWriter, mgoSession *mgo.Session) bool {
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

	module, ok := server.Session.Modules[strings.ToLower(data.Module)]
	if !ok {
		log.Errorf("Unknown module : %s %+v %+v", data.Module, module, ok)
	}

	ip := strings.Split(conn.RemoteAddr().String(), ":")[0]
	projectName, err := server.getProjectByNodeIP(ip)

	log.Debugf("%+v", data)
	err = module.WriteDb(data, mgoSession, projectName)

	if err != nil {
		log.Errorf("Database error : %+v", err)
	}
	return false

}
