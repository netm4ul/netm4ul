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
	"github.com/netm4ul/netm4ul/core/session"
)

var (
	Version string
	//Nodes represent a map to net.Conn
	Nodes map[string]net.Conn
	//SessionServer represent the server side's session. Hold all the modules
	SessionServer *session.Session
)

const (
	//CapacityLow defines the lowest tier for a performance metric
	CapacityLow = 1
	//CapacityMedium defines the middle tier for a performance metric
	CapacityMedium = 2
	//CapacityHigh defines the highest tier for a performance metric
	CapacityHigh = 3
)

//Requirements defines all the specification needed for a node to be eligble at executing on command.
type Requirements struct {
	NetworkType        string `json:"networktype"`        // "external", "internal", ""
	ConnectionCapacity uint16 `json:"connectioncapacity"` // CapacityLow, CapacityMedium, CapacityHigh
	ComputingCapacity  uint16 `json:"computingcapacity"`  // CapacityLow, CapacityMedium, CapacityHigh
}

//Command represents the communication protocol between clients and the master node
type Command struct {
	Name         string       `json:"name"`
	Options      []string     `json:"options"`
	Requirements Requirements `json:"requirements"`
}

func init() {
	Nodes = make(map[string]net.Conn)
}

// CreateServer : Initialise the infinite server loop on the master node
func CreateServer(s *session.Session) {
	SessionServer = s
	Listen(s)
}

// Listen : create the TCP server
func Listen(s *session.Session) {

	Version = s.Config.Versions.Server
	ipport := s.GetServerIPPort()

	var err error
	var l net.Listener

	if s.Config.TLSParams.UseTLS {
		l, err = tls.Listen("tcp", ipport, s.Config.TLSParams.TLSConfig)

	} else {
		l, err = net.Listen("tcp", ipport)
	}

	if err != nil {
		log.Fatalf("Error listening : %s", err.Error())
	}

	// Close the listener when the application closes.
	defer l.Close()

	database.InitDatabase(&s.Config)
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
		go handleRequest(conn, mgoSessionClone, s)
	}
}

func handleRequest(conn net.Conn, mgoSession *mgo.Session, s *session.Session) {

	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	// defer conn.Close()

	handleHello(conn, rw, mgoSession, s)

	stop := false
	for !stop {
		stop = handleData(conn, rw, mgoSession, s)
	}
}

// Recv basic info for the node at connection time.
func handleHello(conn net.Conn, rw *bufio.ReadWriter, mgoSession *mgo.Session, s *session.Session) {

	var node config.Node

	dec := gob.NewDecoder(rw)
	err := dec.Decode(&node)

	if err != nil {
		log.Errorf("Cannot read hello data : %s", err.Error())
		return
	}

	ip := strings.Split(conn.RemoteAddr().String(), ":")[0]

	if s.Config.Verbose {
		if _, ok := s.Config.Nodes[ip]; ok {
			log.Infoln("Node known. Updating")
		} else {
			log.Infoln("Unknown node. Creating")
		}
	}

	s.Config.Nodes = make(map[string]config.Node)
	s.Config.Modules = make(map[string]config.Module)

	s.Config.Nodes[ip] = node
	Nodes[ip] = conn
	database.CreateProject(mgoSession, node.Project)

	p := database.GetProjects(mgoSession)

	log.Debugf("Nodes : %+v", s.Config.Nodes)
	log.Debugf("Projects : %+v", p)

}

func getProjectByNodeIP(ip string, s *session.Session) (string, error) {

	var err error

	n, ok := s.Config.Nodes[ip]
	if !ok {
		return "", errors.New("Unknown node ! Could not get project name")
	}

	project := n.Project

	return project, err
}

//SendCmdByName is a wrapper to the SendCommand function.
func SendCmdByName(name string, options []string) {
	//TODO get the Command by module name & setup options and requirements

	// cmd := Command{
	// 	Name: name,
	// 	Options: options,
	// 	Requirements: GetRequirementFromCommandName(name)
	// }
	// SendCmd(cmd)
}

//SendCmd sends one commands with its options to selected clients
func SendCmd(command Command, s *session.Session) error {

	conns, err := getAvailableNodes(command.Requirements)

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
func getAvailableNodes(req Requirements) ([]net.Conn, error) {
	// TODO : Requirements for each modules and load balance
	var availables []net.Conn
	for _, conn := range Nodes {
		availables = append(availables, conn)
	}
	return availables, nil
}

// handleData decode and route all data after the "hello". It listens forever until connection closed.
func handleData(conn net.Conn, rw *bufio.ReadWriter, mgoSession *mgo.Session, s *session.Session) bool {
	var data modules.Result
	var err error

	dec := gob.NewDecoder(rw)
	err = dec.Decode(&data)

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

	module, ok := s.Modules[strings.ToLower(data.Module)]
	if !ok {
		log.Errorf("Unknown module : %s %+v %+v", data.Module, module, ok)
	}

	ip := strings.Split(conn.RemoteAddr().String(), ":")[0]
	projectName, err := getProjectByNodeIP(ip, s)

	log.Debugf("%+v", data)
	err = module.WriteDb(data, mgoSession, projectName)

	if err != nil {
		log.Errorf("Database error : %+v", err)
	}
	return false

}
