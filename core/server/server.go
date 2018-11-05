package server

import (
	"bufio"
	"crypto/tls"
	"encoding/gob"
	"errors"
	"io"
	"net"
	"strings"

	"github.com/netm4ul/netm4ul/core/communication"
	"github.com/netm4ul/netm4ul/core/config"
	"github.com/netm4ul/netm4ul/core/database"
	"github.com/netm4ul/netm4ul/core/database/models"
	"github.com/netm4ul/netm4ul/core/events"
	"github.com/netm4ul/netm4ul/core/session"
	log "github.com/sirupsen/logrus"
)

//Server represent the base structure for the server node.
type Server struct {
	//Session represent the server side's session. Hold all the modules
	Session *session.Session
	Db      models.Database
}

// CreateServer : Initialise the infinite server loop on the master node
func CreateServer(s *session.Session) *Server {
	server := Server{Session: s}
	db, err := database.NewDatabase(&server.Session.Config)
	if err != nil || db == nil {
		panic(err)
	}
	server.Db = db

	server.Session.Nodes = make(map[string]communication.Node, 0)
	server.Session.Config.Modules = make(map[string]config.Module)

	go server.SetupEventsPropagations()
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

//SetupEventsPropagations will listen on the EventsQueue.
func (server *Server) SetupEventsPropagations() {
	for {
		ev := <-events.EventQueue
		server.RunModulesByEvents(ev)
	}
}

func (server *Server) RunModulesByEvents(ev events.Event) {
	for moduleName, module := range server.Session.ModulesEnabled {
		requiredEv := module.DependsOn()
		if requiredEv == ev.Type {
			log.Printf("Module %s will be run (event of type : %s received)\n", moduleName, ev.Type)
			//TODO : actually run the modules !
		}
	}
}

func (server *Server) handleRequest(conn net.Conn) {

	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	log.Debug("Modules : " + server.Session.GetModulesList())
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

	server.Session.Nodes[node.ID] = node

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

	server.Session.Algo.SetNodes(server.Session.Nodes)
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
func (server *Server) getNextNodes(cmd communication.Command) (map[string]communication.Node, error) {
	// TODO : Requirements for each modules and load balance
	nodes := server.Session.Algo.NextExecutionNodes(cmd)
	log.Debug("Selected nodes : ", nodes)

	return nodes, nil
}

// handleData decode and route all data after the "hello". It listens forever until connection closed.
// the boolean is a "stop" boolean. True indicate stop connection
func (server *Server) handleData(conn net.Conn, rw *bufio.ReadWriter) bool {
	var data communication.Result

	err := gob.NewDecoder(rw).Decode(&data)

	// handle connection closed (client shutdown). And remove it from the server.Session.Nodes slice !
	if err == io.EOF {
		log.Errorf("Connection closed : %s [%s]", err.Error(), conn.RemoteAddr())
		for i, node := range server.Session.Nodes {
			if node.Conn.RemoteAddr() == conn.RemoteAddr() {
				delete(server.Session.Nodes, i)
			}
		}
		// stop all handleData for this conn
		return true
	}

	// handle other error
	if err != nil {
		log.Errorf("Error while decoding data : %s", err.Error())
		return false
	}
	log.Debugf("%+v", data)

	module, ok := server.Session.Modules[strings.ToLower(data.ModuleName)]
	if !ok {
		log.Errorf("Unknown module : %s %+v %+v", data.ModuleName, module, ok)
		return false
	}

	ip := strings.Split(conn.RemoteAddr().String(), ":")[0]

	found := false
	var node communication.Node
	var id string
	for id, node = range server.Session.Nodes {
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
	p := models.Project{
		Name: server.Session.Nodes[id].Project,
	}
	err = server.Db.CreateOrUpdateProject(p)
	if err != nil {
		log.Errorf("Could not create or update project : %+v", err)
		return false
	}

	err = module.WriteDb(data, server.Db, p.Name)

	if err != nil {
		log.Errorf("Database error : %+v", err)
		return false
	}
	log.Infof("Saved database info, module : %s", data.ModuleName)
	node.IsAvailable = true
	return false
}
